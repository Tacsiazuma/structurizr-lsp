package lsp

import (
	"fmt"
	"log"

	"github.com/tacsiazuma/structurizr-lsp/parser"
	"github.com/tacsiazuma/structurizr-lsp/rpc"
)

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

var inlayTokens = []parser.TokenType{
	parser.TokenName, parser.TokenDescription, parser.TokenValue,
}

func (l *Lsp) findInlayHints(node *parser.ASTNode, rng Range) []InlayHint {
	hints := make([]InlayHint, 0)
	for _, v := range inlayTokens {
		if v == node.Token.Type {
			hints = append(hints, InlayHint{
				Label:    fmt.Sprintf("%s: ", node.Token.Type),
				Position: Position{Line: node.Location.Line, Character: node.Location.Pos},
			})
		}
	}
	for _, attribute := range node.Attributes {
		for _, v := range inlayTokens {
			if v == attribute.Type {
				hints = append(hints, InlayHint{
					Label:    fmt.Sprintf("%s: ", attribute.Type),
					Position: Position{Line: attribute.Location.Line, Character: attribute.Location.Pos},
				})
			}
		}
	}
	for _, child := range node.Children {
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
