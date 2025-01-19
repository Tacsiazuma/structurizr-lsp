package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"runtime/debug"
	"strings"

	"github.com/tacsiazuma/structurizr-lsp/parser"
)

func main() {
	initLogger()
	if os.Args[len(os.Args)-1] == "version" {
		info, _ := debug.ReadBuildInfo()
		fmt.Println(info.Main.Version)
		return
	}
	// Defer a function to recover from panics
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Recovered from panic: %v\n", r)
		}
	}()
	for {
		// Read message from client
		msg, err := readMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read message: %v\n", err)
			continue
		}

		// Parse the JSON-RPC request
		var req Request
		if err := json.Unmarshal([]byte(msg), &req); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse JSON: %v\n", err)
			continue
		}

		// Handle the request
		switch req.Method {
		case "initialize":
			handleInitialize(req)
		case "initialized":
			break
		case "textDocument/didSave":
			break
		case "textDocument/didChange":
			var params DidChangeTextDocumentParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				fmt.Printf("Failed to parse 'didOpen' params: %v\n", err)
				return
			}
			handleDidChange(params)
		case "textDocument/didOpen":
			var params DidOpenTextDocumentParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				fmt.Printf("Failed to parse 'didOpen' params: %v\n", err)
				return
			}
			handleDidOpen(params)
		case "shutdown":
			handleShutdown(req)
		default:
			sendError(req.ID, -32601, "Method not found")
		}
	}
}

func handleDidOpen(param DidOpenTextDocumentParams) {
	p := parser.New(strings.TrimPrefix(param.TextDocument.URI, "file://"), param.TextDocument.Text, parser.NewIncluder())
	_, diags := p.Parse()
	publishDiagnostics(diags)
}

func publishDiagnostics(diags []*parser.Diagnostic) {
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
		notification := Notification{
			Method: "textDocument/publishDiagnostics",
			Params: params,
		}
		writeNotification(notification)
	}
}

func handleDidChange(param DidChangeTextDocumentParams) {
	p := parser.New(strings.TrimPrefix(param.TextDocument.URI, "file://"), param.ContentChanges[0].Text, parser.NewIncluder())
	_, diags := p.Parse()
	publishDiagnostics(diags)
}

func handleInitialize(req Request) {
	// Respond with basic server capabilities
	capabilities := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"textDocumentSync": 1,
			"completionProvider": map[string]bool{
				"resolveProvider": true,
			},
		},
	}
	response := Response{
		Jsonrpc: "2.0",
		ID:      req.ID,
		Result:  capabilities,
	}
	if err := writeMessage(response); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send response: %v\n", err)
	}
}

func handleShutdown(req Request) {
	response := Response{
		Jsonrpc: "2.0",
		ID:      req.ID,
		Result:  nil,
	}
	if err := writeMessage(response); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send response: %v\n", err)
	}
	os.Exit(0)
}

func sendError(id int, code int, message string) {
	response := Response{
		Jsonrpc: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
	if err := writeMessage(response); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send error response: %v\n", err)
	}
}
