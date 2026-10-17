package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// StdioTransport implements MCP over stdin/stdout.
type StdioTransport struct {
	mu       sync.Mutex
	server   *Server
	reader   *bufio.Reader
	writer   *bufio.Writer
	out      io.Writer
}

// NewStdioTransport creates a stdio transport.
func NewStdioTransport(server *Server) *StdioTransport {
	return &StdioTransport{
		server: server,
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
		out:    os.Stdout,
	}
}

// Run starts the stdio transport loop.
func (t *StdioTransport) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := t.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		line = line[:len(line)-1] // trim newline
		if len(line) == 0 {
			continue
		}

		var msg JSONRPCMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			resp := NewErrorResponse(nil, -32700, fmt.Sprintf("Parse error: %v", err))
			_ = t.Send(resp)
			continue
		}

		resp := t.server.HandleMessage(msg)
		// Don't send response for notifications (no id)
		if msg.ID != nil || resp.Error != nil {
			if err := t.Send(resp); err != nil {
				return err
			}
		}
	}
}

// Send writes a JSON-RPC message to stdout.
func (t *StdioTransport) Send(msg JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err := t.writer.Write(data); err != nil {
		return err
	}
	if _, err := t.writer.WriteString("\n"); err != nil {
		return err
	}
	return t.writer.Flush()
}
