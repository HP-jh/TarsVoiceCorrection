package mcp

import (
	"context"
	"io"
)

// Transport is the interface for MCP transports.
type Transport interface {
	// Run starts the transport and blocks until ctx is cancelled or an error occurs.
	Run(ctx context.Context) error
	// Send sends a JSON-RPC message.
	Send(msg JSONRPCMessage) error
}

// JSONRPCMessage represents a JSON-RPC 2.0 message.
type JSONRPCMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError is a JSON-RPC error object.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MessageReader reads JSON-RPC messages.
type MessageReader interface {
	ReadMessage() (*JSONRPCMessage, error)
}

// MessageWriter writes JSON-RPC messages.
type MessageWriter interface {
	WriteMessage(msg *JSONRPCMessage) error
}

// ReadWriteTransport combines reader and writer.
type ReadWriteTransport interface {
	MessageReader
	MessageWriter
	io.Closer
}

const JSONRPCVersion = "2.0"

// NewSuccessResponse creates a successful JSON-RPC response.
func NewSuccessResponse(id interface{}, result interface{}) JSONRPCMessage {
	return JSONRPCMessage{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse creates an error JSON-RPC response.
func NewErrorResponse(id interface{}, code int, message string) JSONRPCMessage {
	return JSONRPCMessage{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}
