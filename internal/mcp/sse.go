package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// SSETransport implements MCP over Server-Sent Events.
type SSETransport struct {
	mu       sync.RWMutex
	server   *Server
	addr     string
	sessions map[string]*sseSession
	srv      *http.Server
}

type sseSession struct {
	id       string
	messages chan JSONRPCMessage
}

// NewSSETransport creates an SSE transport.
func NewSSETransport(server *Server, addr string) *SSETransport {
	return &SSETransport{
		server:   server,
		addr:     addr,
		sessions: make(map[string]*sseSession),
	}
}

// Run starts the SSE HTTP server.
func (t *SSETransport) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", t.handleSSE)
	mux.HandleFunc("/message", t.handleMessage)
	mux.HandleFunc("/health", t.handleHealth)

	t.srv = &http.Server{
		Addr:    t.addr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- t.srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = t.srv.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (t *SSETransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	sessionID := fmt.Sprintf("sess-%d", time.Now().UnixNano())
	session := &sseSession{
		id:       sessionID,
		messages: make(chan JSONRPCMessage, 100),
	}

	t.mu.Lock()
	t.sessions[sessionID] = session
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.sessions, sessionID)
		t.mu.Unlock()
		close(session.messages)
	}()

	// Send endpoint URL as first event
	fmt.Fprintf(w, "event: endpoint\ndata: /message?session_id=%s\n\n", sessionID)
	flusher.Flush()

	for {
		select {
		case msg, ok := <-session.messages:
			if !ok {
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(data))
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (t *SSETransport) handleMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}

	t.mu.RLock()
	session, ok := t.sessions[sessionID]
	t.mu.RUnlock()
	if !ok {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	var msg JSONRPCMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	resp := t.server.HandleMessage(msg)
	if msg.ID != nil || resp.Error != nil {
		select {
		case session.messages <- resp:
		default:
			log.Printf("SSE session %s message buffer full", sessionID)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (t *SSETransport) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Send is a no-op for SSE (messages are sent via HTTP POST -> SSE channel).
func (t *SSETransport) Send(msg JSONRPCMessage) error {
	return nil
}
