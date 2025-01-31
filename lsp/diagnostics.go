package lsp

import (
	"fmt"
	"net/url"
	"os"

	"github.com/tacsiazuma/structurizr-lsp/parser"
	"github.com/tacsiazuma/structurizr-lsp/rpc"
)

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
		if err := l.rpc.WriteMessage(notification); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send error response: %v\n", err)
		}
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
	if err := l.rpc.WriteMessage(notification); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send error response: %v\n", err)
	}
}
