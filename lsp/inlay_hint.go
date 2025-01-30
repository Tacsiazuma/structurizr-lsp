package lsp

import (
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

