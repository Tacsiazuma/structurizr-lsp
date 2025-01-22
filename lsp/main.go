package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/tacsiazuma/structurizr-lsp/parser"
	"github.com/tacsiazuma/structurizr-lsp/rpc"
)

type Diagnostic struct {
	Range   Range  `json:"range"`
	Message string `json:"message"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type PublishDiagnosticsParams struct {
	URI         string        `json:"uri"`
	Diagnostics []*Diagnostic `json:"diagnostics"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   TextDocumentItem `json:"textDocument"`
	ContentChanges []ContentChange
}

type ContentChange struct {
	Text string `json:"text"`
}

func (l *Lsp) handleDidOpen(param DidOpenTextDocumentParams) {
	a := parser.NewAnalyser(strings.TrimPrefix(param.TextDocument.URI, "file://"), param.TextDocument.Text)
	_, _, diags := a.Analyse()
	if len(diags) == 0 {
		l.clearDiagnostics(param.TextDocument.URI)
	} else {
		l.publishDiagnostics(diags)
	}
}

func (l *Lsp) publishDiagnostics(diags []*parser.Diagnostic) {
	diagnostics := make(map[string][]*Diagnostic, 0)
	for _, diag := range diags {
		diagnostics[diag.Location.Source] = append(diagnostics[diag.Location.Source], &Diagnostic{
			Message: diag.Message,
			Range:   Range{Start: Position{Character: diag.Location.Pos, Line: diag.Location.Line}, End: Position{Character: diag.Location.Pos, Line: diag.Location.Line}}})
	}
	for k, v := range diagnostics {
		uri := &url.URL{
			Scheme: "file",
			Path:   k,
		}
		params := PublishDiagnosticsParams{
			URI:         uri.String(),
			Diagnostics: v,
		}
		notification := rpc.Notification{
			Method: "textDocument/publishDiagnostics",
			Params: params,
		}
		l.rpc.WriteMessage(notification)
	}
}

func (l *Lsp) handleDidChange(param DidChangeTextDocumentParams) {
	p := parser.NewAnalyser(strings.TrimPrefix(param.TextDocument.URI, "file://"), param.ContentChanges[0].Text)
	_, _, diags := p.Analyse()
	if len(diags) == 0 {
		l.clearDiagnostics(param.TextDocument.URI)
	} else {
		l.publishDiagnostics(diags)
	}
}

func (l *Lsp) clearDiagnostics(s string) {
	params := PublishDiagnosticsParams{
		URI:         s,
		Diagnostics: make([]*Diagnostic, 0),
	}
	notification := rpc.Notification{
		Method: "textDocument/publishDiagnostics",
		Params: params,
	}
	l.rpc.WriteMessage(notification)
}

func (l *Lsp) handleInitialize(req rpc.Request) {
	l.initialized = true
	// Respond with basic server capabilities
	capabilities := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"textDocumentSync": 1,
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
	case "textDocument/completion": // not implemented yet
		break
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

type Lsp struct {
	initialized bool
	rpc         *rpc.Rpc
}

func From(input io.Reader, output io.Writer) *Lsp {
	r := rpc.NewRpc(input, output)
	return &Lsp{rpc: r}
}

