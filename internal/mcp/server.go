package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/HP-jh/TarsVoiceCorrection/internal/corrector"
)

// Server implements an MCP server with voice correction tools.
type Server struct {
	mu        sync.RWMutex
	corr      *corrector.Corrector
	transport Transport
	tools     []Tool
}

// Tool represents an MCP tool.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema describes tool input parameters.
type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]Property    `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// Property describes a single input property.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// NewServer creates a new MCP server.
func NewServer(corr *corrector.Corrector) *Server {
	s := &Server{corr: corr}
	s.initTools()
	return s
}

func (s *Server) initTools() {
	s.tools = []Tool{
		{
			Name:        "voice_correction",
			Description: "Correct voice input text: spelling errors, homophones, synonyms, mixed CJK-English. Returns corrected text with confidence and candidates.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"text": {Type: "string", Description: "The input text to correct"},
				},
				Required: []string{"text"},
			},
		},
		{
			Name:        "pinyin_match",
			Description: "Find homophones (same or similar pinyin) for a given Chinese word or phrase.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"word":  {Type: "string", Description: "The word to find homophones for"},
					"top_k": {Type: "integer", Description: "Maximum number of results (default 5)"},
				},
				Required: []string{"word"},
			},
		},
		{
			Name:        "synonym_lookup",
			Description: "Look up synonyms for a word using the built-in dictionary.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"word": {Type: "string", Description: "The word to look up synonyms for"},
				},
				Required: []string{"word"},
			},
		},
		{
			Name:        "fuzzy_search",
			Description: "Fuzzy search for similar words in the dictionary using edit distance and n-gram similarity.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"query":   {Type: "string", Description: "The query text"},
					"top_k":   {Type: "integer", Description: "Maximum number of results (default 5)"},
					"threshold": {Type: "number", Description: "Minimum similarity threshold 0-1 (default 0.5)"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "lang_detect",
			Description: "Detect the dominant language of input text (zh/en/mixed/numeric/unknown).",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"text": {Type: "string", Description: "The text to analyze"},
				},
				Required: []string{"text"},
			},
		},
	}
}

// SetTransport sets the transport for the server.
func (s *Server) SetTransport(t Transport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transport = t
}

// Run starts the server with the given transport.
func (s *Server) Run(ctx context.Context) error {
	s.mu.RLock()
	t := s.transport
	s.mu.RUnlock()
	if t == nil {
		return fmt.Errorf("no transport set")
	}
	return t.Run(ctx)
}

// HandleMessage processes a single JSON-RPC message and returns the response.
func (s *Server) HandleMessage(msg JSONRPCMessage) JSONRPCMessage {
	if msg.JSONRPC != JSONRPCVersion && msg.JSONRPC != "" {
		return NewErrorResponse(msg.ID, -32600, "Invalid Request: invalid jsonrpc version")
	}

	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return JSONRPCMessage{JSONRPC: JSONRPCVersion} // notification, no response
	case "tools/list":
		return s.handleToolsList(msg)
	case "tools/call":
		return s.handleToolCall(msg)
	case "":
		// Response message (server shouldn't receive these normally)
		return JSONRPCMessage{JSONRPC: JSONRPCVersion}
	default:
		return NewErrorResponse(msg.ID, -32601, fmt.Sprintf("Method not found: %s", msg.Method))
	}
}

func (s *Server) handleInitialize(msg JSONRPCMessage) JSONRPCMessage {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "tars-voice-correction",
			"version": "0.1.0",
		},
	}
	return NewSuccessResponse(msg.ID, result)
}

func (s *Server) handleToolsList(msg JSONRPCMessage) JSONRPCMessage {
	result := map[string]interface{}{
		"tools": s.tools,
	}
	return NewSuccessResponse(msg.ID, result)
}

func (s *Server) handleToolCall(msg JSONRPCMessage) JSONRPCMessage {
	paramsBytes, err := json.Marshal(msg.Params)
	if err != nil {
		return NewErrorResponse(msg.ID, -32602, "Invalid params")
	}

	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return NewErrorResponse(msg.ID, -32602, "Invalid params")
	}

	var result interface{}
	var callErr error

	switch params.Name {
	case "voice_correction":
		result, callErr = s.callVoiceCorrection(params.Arguments)
	case "pinyin_match":
		result, callErr = s.callPinyinMatch(params.Arguments)
	case "synonym_lookup":
		result, callErr = s.callSynonymLookup(params.Arguments)
	case "fuzzy_search":
		result, callErr = s.callFuzzySearch(params.Arguments)
	case "lang_detect":
		result, callErr = s.callLangDetect(params.Arguments)
	default:
		return NewErrorResponse(msg.ID, -32601, fmt.Sprintf("Unknown tool: %s", params.Name))
	}

	if callErr != nil {
		return NewErrorResponse(msg.ID, -32000, callErr.Error())
	}

	content := []map[string]interface{}{
		{
			"type": "text",
			"text": toJSONString(result),
		},
	}
	return NewSuccessResponse(msg.ID, map[string]interface{}{"content": content})
}

func (s *Server) callVoiceCorrection(args json.RawMessage) (interface{}, error) {
	var p struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return nil, err
	}
	if strings.TrimSpace(p.Text) == "" {
		return nil, fmt.Errorf("text is required")
	}
	result := s.corr.Correct(p.Text)
	return result, nil
}

func (s *Server) callPinyinMatch(args json.RawMessage) (interface{}, error) {
	var p struct {
		Word string `json:"word"`
		TopK int    `json:"top_k"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return nil, err
	}
	if p.TopK <= 0 {
		p.TopK = 5
	}
	matches := s.corr.PinyinMatch(p.Word, p.TopK)
	return map[string]interface{}{
		"word":    p.Word,
		"matches": matches,
	}, nil
}

func (s *Server) callSynonymLookup(args json.RawMessage) (interface{}, error) {
	var p struct {
		Word string `json:"word"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return nil, err
	}
	syns := s.corr.SynonymLookup(p.Word)
	return map[string]interface{}{
		"word":     p.Word,
		"synonyms": syns,
	}, nil
}

func (s *Server) callFuzzySearch(args json.RawMessage) (interface{}, error) {
	var p struct {
		Query     string  `json:"query"`
		TopK      int     `json:"top_k"`
		Threshold float64 `json:"threshold"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return nil, err
	}
	if p.TopK <= 0 {
		p.TopK = 5
	}
	if p.Threshold <= 0 {
		p.Threshold = 0.5
	}
	matches := s.corr.FuzzySearch(p.Query, p.TopK)
	return map[string]interface{}{
		"query":   p.Query,
		"matches": matches,
	}, nil
}

func (s *Server) callLangDetect(args json.RawMessage) (interface{}, error) {
	var p struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return nil, err
	}
	lang := s.corr.LangDetect(p.Text)
	return map[string]interface{}{
		"text":     p.Text,
		"language": lang,
	}, nil
}

func toJSONString(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
