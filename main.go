package main

import (
	"fmt"
	"os"
	"tacsiazuma/structurizr-lsp/parser"
)

func main() {
	content, _ := os.ReadFile("test.dsl")
	tokens, _ := parser.Lexer(string(content))
	for _, token := range tokens {
        fmt.Printf("%d:%d Type: %s Content: %s\n", token.Location.Line, token.Location.Pos, token.Type, token.Content)
	}
}
