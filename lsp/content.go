package lsp

import (
	"fmt"
	"strings"

	"github.com/tacsiazuma/structurizr-lsp/parser"
)

func (l *Lsp) registerContent(uri, content string, ast *parser.ASTNode) {
	l.content[uri] = Content{Text: content, Ast: ast}
	l.logger.Println("Writing " + uri)
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

