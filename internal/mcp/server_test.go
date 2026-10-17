package mcp

import (
	"encoding/json"
	"testing"

	"github.com/HP-jh/TarsVoiceCorrection/internal/corrector"
)

func TestServerHandleInitialize(t *testing.T) {
	corr, err := corrector.NewCorrector()
	if err != nil {
		t.Fatalf("NewCorrector: %v", err)
	}
	server := NewServer(corr)

	msg := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	resp := server.HandleMessage(msg)
	if resp.Error != nil {
		t.Fatalf("initialize error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}
	if result["protocolVersion"] == nil {
		t.Error("Expected protocolVersion in result")
	}
}

func TestServerHandleToolsList(t *testing.T) {
	corr, _ := corrector.NewCorrector()
	server := NewServer(corr)

	msg := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	resp := server.HandleMessage(msg)
	if resp.Error != nil {
		t.Fatalf("tools/list error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}
	tools, ok := result["tools"].([]Tool)
	if !ok {
		t.Fatal("Expected tools array")
	}
	if len(tools) != 5 {
		t.Errorf("Expected 5 tools, got %d", len(tools))
	}
}

func TestServerHandleToolCallVoiceCorrection(t *testing.T) {
	corr, _ := corrector.NewCorrector()
	server := NewServer(corr)

	params := map[string]interface{}{
		"name": "voice_correction",
		"arguments": map[string]string{
			"text": "电恼",
		},
	}
	paramsBytes, _ := json.Marshal(params)
	var paramsRaw interface{}
	json.Unmarshal(paramsBytes, &paramsRaw)

	msg := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  paramsRaw,
	}
	resp := server.HandleMessage(msg)
	if resp.Error != nil {
		t.Fatalf("tools/call error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}
	content, ok := result["content"].([]map[string]interface{})
	if !ok || len(content) == 0 {
		t.Fatal("Expected content array")
	}
	text, ok := content[0]["text"].(string)
	if !ok || text == "" {
		t.Error("Expected non-empty text content")
	}
}

func TestServerHandleToolCallPinyinMatch(t *testing.T) {
	corr, _ := corrector.NewCorrector()
	server := NewServer(corr)

	params := map[string]interface{}{
		"name": "pinyin_match",
		"arguments": map[string]interface{}{
			"word":  "中国",
			"top_k": 3,
		},
	}
	paramsBytes, _ := json.Marshal(params)
	var paramsRaw interface{}
	json.Unmarshal(paramsBytes, &paramsRaw)

	msg := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  paramsRaw,
	}
	resp := server.HandleMessage(msg)
	if resp.Error != nil {
		t.Fatalf("tools/call error: %v", resp.Error)
	}
}

func TestServerHandleToolCallSynonymLookup(t *testing.T) {
	corr, _ := corrector.NewCorrector()
	server := NewServer(corr)

	params := map[string]interface{}{
		"name": "synonym_lookup",
		"arguments": map[string]string{
			"word": "电脑",
		},
	}
	paramsBytes, _ := json.Marshal(params)
	var paramsRaw interface{}
	json.Unmarshal(paramsBytes, &paramsRaw)

	msg := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "tools/call",
		Params:  paramsRaw,
	}
	resp := server.HandleMessage(msg)
	if resp.Error != nil {
		t.Fatalf("tools/call error: %v", resp.Error)
	}
}

func TestServerHandleToolCallFuzzySearch(t *testing.T) {
	corr, _ := corrector.NewCorrector()
	server := NewServer(corr)

	params := map[string]interface{}{
		"name": "fuzzy_search",
		"arguments": map[string]interface{}{
			"query":     "电恼",
			"top_k":     3,
			"threshold": 0.3,
		},
	}
	paramsBytes, _ := json.Marshal(params)
	var paramsRaw interface{}
	json.Unmarshal(paramsBytes, &paramsRaw)

	msg := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "tools/call",
		Params:  paramsRaw,
	}
	resp := server.HandleMessage(msg)
	if resp.Error != nil {
		t.Fatalf("tools/call error: %v", resp.Error)
	}
}

func TestServerHandleToolCallLangDetect(t *testing.T) {
	corr, _ := corrector.NewCorrector()
	server := NewServer(corr)

	params := map[string]interface{}{
		"name": "lang_detect",
		"arguments": map[string]string{
			"text": "Hello 世界",
		},
	}
	paramsBytes, _ := json.Marshal(params)
	var paramsRaw interface{}
	json.Unmarshal(paramsBytes, &paramsRaw)

	msg := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      7,
		Method:  "tools/call",
		Params:  paramsRaw,
	}
	resp := server.HandleMessage(msg)
	if resp.Error != nil {
		t.Fatalf("tools/call error: %v", resp.Error)
	}
}

func TestServerHandleUnknownMethod(t *testing.T) {
	corr, _ := corrector.NewCorrector()
	server := NewServer(corr)

	msg := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      99,
		Method:  "unknown/method",
	}
	resp := server.HandleMessage(msg)
	if resp.Error == nil {
		t.Fatal("Expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("Expected code -32601, got %d", resp.Error.Code)
	}
}

func TestJSONRPCHelpers(t *testing.T) {
	resp := NewSuccessResponse(1, map[string]string{"status": "ok"})
	if resp.ID != 1 || resp.Result == nil {
		t.Error("NewSuccessResponse failed")
	}

	errResp := NewErrorResponse(2, -32000, "test error")
	if errResp.Error == nil || errResp.Error.Message != "test error" {
		t.Error("NewErrorResponse failed")
	}
}
