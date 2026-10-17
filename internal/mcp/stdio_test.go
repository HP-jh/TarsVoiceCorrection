package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/HP-jh/TarsVoiceCorrection/internal/corrector"
)

func TestStdioTransportRoundTrip(t *testing.T) {
	corr, _ := corrector.NewCorrector()
	server := NewServer(corr)

	inBuf := &bytes.Buffer{}
	outBuf := &bytes.Buffer{}

	trans := &StdioTransport{
		server: server,
		reader: bufio.NewReader(inBuf),
		writer: bufio.NewWriter(outBuf),
		out:    outBuf,
	}

	// Write an initialize request
	req := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	reqBytes, _ := json.Marshal(req)
	inBuf.Write(reqBytes)
	inBuf.WriteByte('\n')

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run in background
	done := make(chan error, 1)
	go func() {
		done <- trans.Run(ctx)
	}()

	// Give it time to process
	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	outStr := outBuf.String()
	if !strings.Contains(outStr, `"result"`) {
		t.Fatalf("Expected result in output, got: %s", outStr)
	}
}

func TestStdioTransportSend(t *testing.T) {
	outBuf := &bytes.Buffer{}
	trans := &StdioTransport{
		writer: bufio.NewWriter(outBuf),
		out:    outBuf,
	}

	msg := NewSuccessResponse(1, "ok")
	if err := trans.Send(msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	outStr := outBuf.String()
	if !strings.Contains(outStr, `"jsonrpc"`) {
		t.Fatalf("Expected jsonrpc in output, got: %s", outStr)
	}
}
