package lsp

import (
	"strings"

	"github.com/tacsiazuma/structurizr-lsp/parser"
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
