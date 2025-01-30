package lsp

import (
	"fmt"
	"os"

	"github.com/tacsiazuma/structurizr-lsp/rpc"
)

func (l *Lsp) handleInitialize(req rpc.Request) {
	l.initialized = true
	// Respond with basic server capabilities
	capabilities := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"textDocumentSync":           1,
			"documentFormattingProvider": true,
			"inlayHintProvider":          true,
			"completionProvider": map[string]bool{
				"resolveProvider": true,
			},
		},
	}
	response := rpc.Response{
		Jsonrpc: "2.0",
		ID:      req.ID,
		Result:  capabilities,
	}
	if err := l.rpc.WriteMessage(response); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send response: %v\n", err)
	}
}

func (l *Lsp) handleShutdown(req rpc.Request) {
	response := rpc.Response{
		Jsonrpc: "2.0",
		ID:      req.ID,
		Result:  nil,
	}
	if err := l.rpc.WriteMessage(response); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send response: %v\n", err)
	}
	os.Exit(0)
}

