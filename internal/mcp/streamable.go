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

// StreamableTransport implements MCP over HTTP with JSON streaming.
type StreamableTransport struct {
	mu     sync.Mutex
	server *Server
	addr   string
	srv    *http.Server
}

// NewStreamableTransport creates a streamable HTTP transport.
func NewStreamableTransport(server *Server, addr string) *StreamableTransport {
	return &StreamableTransport{
		server: server,
		addr:   addr,
	}
}

// Run starts the streamable HTTP server.
func (t *StreamableTransport) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", t.handleMCP)
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

func (t *StreamableTransport) handleMCP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var msg JSONRPCMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Invalid JSON: %v"}`, err), http.StatusBadRequest)
		return
	}

	resp := t.server.HandleMessage(msg)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("streamable encode error: %v", err)
	}
}

func (t *StreamableTransport) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "transport": "streamable"})
}

// Send is a no-op for streamable transport.
func (t *StreamableTransport) Send(msg JSONRPCMessage) error {
	return nil
}
