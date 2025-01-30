package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/tacsiazuma/structurizr-lsp/parser"
	"github.com/tacsiazuma/structurizr-lsp/rpc"
)

type Lsp struct {
	initialized bool
	rpc         *rpc.Rpc
	logger      *log.Logger
	content     map[string]Content
}

func From(input io.Reader, output io.Writer, logger *log.Logger) *Lsp {
	r := rpc.NewRpc(input, output, logger)
	return &Lsp{rpc: r, logger: logger, content: make(map[string]Content)}
}

func (l *Lsp) sendError(id int, code int, message string) {
	response := rpc.Response{
		Jsonrpc: "2.0",
		ID:      id,
		Error: &rpc.Error{
			Code:    code,
			Message: message,
		},
	}
	if err := l.rpc.WriteMessage(response); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send error response: %v\n", err)
	}
}

func (l *Lsp) Handle() error {
	// Read message from client
	msg, err := l.rpc.ReadMessage()
	if err != nil {
		return fmt.Errorf("Failed to read message: %v", err)
	}

	// Parse the JSON-RPC request
	var req rpc.Request
	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		return fmt.Errorf("Failed to parse JSON: %v", err)
	}

	// Handle the request
	switch req.Method {
	case "initialize":
		l.handleInitialize(req)
	case "initialized": // notification does not require response
		break
	case "textDocument/didSave": // notification does not require response
		break
	case "textDocument/didClose": // notification does not require response
		break
	case "textDocument/formatting": // notification does not require response
		var params FormattingParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return fmt.Errorf("Failed to parse 'inlayHint' params: %v", err)
		}
		l.handleFormatting(req.ID, params)
	case "textDocument/completion": // not implemented yet
		break
	case "textDocument/inlayHint":
		var params InlayHintParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return fmt.Errorf("Failed to parse 'inlayHint' params: %v", err)
		}
		l.handleInlayHint(req.ID, params)
	case "$/cancellation": // not implemented yet
		break
	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return fmt.Errorf("Failed to parse 'didChange' params: %v", err)
		}
		l.handleDidChange(params)
	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return fmt.Errorf("Failed to parse 'didOpen' params: %v", err)
		}
		l.handleDidOpen(params)
	case "shutdown":
		if l.initialized {
			l.handleShutdown(req)
		} else {
			l.sendError(req.ID, -32002, "Not initialized")
		}
	default:
		l.sendError(req.ID, -32601, "Method not found")
	}
	return nil
}

type Content struct {
	Text string
	Ast  *parser.ASTNode
}

