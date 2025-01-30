package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/tacsiazuma/structurizr-lsp/parser"
	"github.com/tacsiazuma/structurizr-lsp/rpc"
)

func (l *Lsp) handleDidOpen(param DidOpenTextDocumentParams) {
	a := parser.NewAnalyser(strings.TrimPrefix(param.TextDocument.URI, "file://"), param.TextDocument.Text)
	_, ast, diags := a.Analyse()
	l.registerContent(param.TextDocument.URI, param.TextDocument.Text, ast)
	if len(diags) == 0 {
		l.clearDiagnostics(param.TextDocument.URI)
	} else {
		l.publishDiagnostics(diags)
	}
}

func (l *Lsp) registerContent(uri, content string, ast *parser.ASTNode) {
	l.content[uri] = Content{Text: content, Ast: ast}
	l.logger.Println("Writing " + uri)
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
	_, ast, diags := p.Analyse()
	l.registerContent(param.TextDocument.URI, param.ContentChanges[0].Text, ast)
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

func (l *Lsp) getContent(uri string) (*Content, error) {
	content, ok := l.content[uri]
	l.logger.Println("Getting " + uri)
	if !ok {
		return nil, fmt.Errorf("Content not found")
	}
	return &content, nil
}

func (l *Lsp) getOrUpdateContent(uri, text string) (*Content, error) {
	if text != "" {
		p := parser.NewAnalyser(strings.TrimPrefix(uri, "file://"), text)
		_, ast, _ := p.Analyse()
		l.registerContent(uri, text, ast)
	}
	content, err := l.getContent(uri)
	return content, err
}

func (l *Lsp) handleInlayHint(id int, param InlayHintParams) {
	content, err := l.getContent(param.TextDocument.URI)
	if err != nil {
		log.Fatal(err)
	}
	hints := l.findInlayHints(content.Ast, param.Range)
	if len(hints) > 0 {
		l.publishInlayHints(id, hints)
	}
}

func (l *Lsp) findInlayHints(ast *parser.ASTNode, rng Range) []InlayHint {
	hints := make([]InlayHint, 0)
	for _, attribute := range ast.Attributes {
		if attribute.Type == parser.TokenName {
			hints = append(hints, InlayHint{
				Label:    "name: ",
				Position: Position{Line: attribute.Location.Line, Character: attribute.Location.Pos},
			})
		}
		if attribute.Type == parser.TokenDescription {
			hints = append(hints, InlayHint{
				Label:    "description: ",
				Position: Position{Line: attribute.Location.Line, Character: attribute.Location.Pos},
			})
		}
	}
	for _, child := range ast.Children {
		hints = append(hints, l.findInlayHints(child, rng)...)
	}
	return hints
}

func (l *Lsp) publishInlayHints(id int, hints []InlayHint) {
	response := rpc.Response{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  hints,
	}
	l.rpc.WriteMessage(response)
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
	case "textDocument/formatting": // notification does not require response
		var params FormattingParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return fmt.Errorf("Failed to parse 'inlayHint' params: %v", err)
		}
		l.handleFormatting(req.ID, params)
	case "textDocument/completion": // not implemented yet
		break
	case "textDocument/inlayHint": // not implemented yet
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

type Lsp struct {
	initialized bool
	rpc         *rpc.Rpc
	logger      *log.Logger
	content     map[string]Content
}

type Content struct {
	Text string
	Ast  *parser.ASTNode
}

func From(input io.Reader, output io.Writer, logger *log.Logger) *Lsp {
	r := rpc.NewRpc(input, output, logger)
	return &Lsp{rpc: r, logger: logger, content: make(map[string]Content)}
}

func (l *Lsp) handleFormatting(id int, param FormattingParams) {
	content, err := l.getOrUpdateContent(param.TextDocument.URI, param.TextDocument.Text)
	if err != nil {
		l.sendError(id, 1, "Cannot format without content")
	}
	input := content.Text
	indentLevel := 0
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))
	var edits []TextEdit
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			lineNum++
			continue // Skip empty lines
		}

		var formattedLine string
		if strings.HasSuffix(line, "{") {
			formattedLine = strings.Repeat("    ", indentLevel) + line
			indentLevel++
		} else if strings.HasPrefix(line, "}") {
			indentLevel--
			if indentLevel < 0 {
				indentLevel = 0 // Prevent negative indentation
			}
			formattedLine = strings.Repeat("    ", indentLevel) + line
		} else {
			formattedLine = strings.Repeat("    ", indentLevel) + line
		}
		sb.WriteString(formattedLine + "\n")
		if formattedLine != line {
			edits = append(edits, TextEdit{
				Range: Range{
					Start: Position{Line: lineNum, Character: 0},
					End:   Position{Line: lineNum, Character: len(line)},
				},
				NewText: formattedLine,
			})
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		l.logger.Println("error reading input: " + err.Error())
		return
	}
	l.registerContent(param.TextDocument.URI, sb.String(), content.Ast) // update the content after formatting
	response := rpc.Response{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  edits,
	}
	l.rpc.WriteMessage(response)
}
