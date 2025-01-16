package parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLexer(t *testing.T) {
	t.Run("should return empty tokens on empty file", func(t *testing.T) {
		content := ""
		tokens, err := Lexer(content)
		if assert.Nil(t, err) {
			assert.Equal(t, make([]Token, 0), tokens)
		}
	})
	t.Run("should return workspace keyword when found", func(t *testing.T) {
		content := "workspace"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenWorkspace, tokens[0].Type)
			assert.Equal(t, "workspace", tokens[0].Content)
		}
	})
	t.Run("should return model keyword when found", func(t *testing.T) {
		content := "model"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenModel, tokens[0].Type)
		}
	})
	t.Run("should return group keyword when found", func(t *testing.T) {
		content := "group"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenGroup, tokens[0].Type)
		}
	})
	t.Run("should return person keyword when found", func(t *testing.T) {
		content := "person"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenPerson, tokens[0].Type)
		}
	})
	t.Run("should return container keyword when found", func(t *testing.T) {
		content := "container"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenContainer, tokens[0].Type)
		}
	})
	t.Run("should return component keyword when found", func(t *testing.T) {
		content := "component"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenComponent, tokens[0].Type)
		}
	})
	t.Run("should return softwareSystem keyword when found", func(t *testing.T) {
		content := "softwareSystem"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenSoftwareSystem, tokens[0].Type)
		}
	})
	t.Run("should return single line comments keyword when found", func(t *testing.T) {
		content := "// comment"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenComment, tokens[0].Type)
			assert.Equal(t, "// comment", tokens[0].Content)
		}
	})
	t.Run("should return open brace symbol when found", func(t *testing.T) {
		content := "{"
		tokens, _ := Lexer(content)
		assert.Equal(t, TokenBraceOpen, tokens[0].Type)
	})
	t.Run("should return close brace symbol when found", func(t *testing.T) {
		content := "}"
		tokens, _ := Lexer(content)
		assert.Equal(t, TokenBraceClose, tokens[0].Type)
	})
	t.Run("should return views keyword when found", func(t *testing.T) {
		content := "views"
		tokens, _ := Lexer(content)
		assert.Equal(t, TokenViews, tokens[0].Type)
	})

	t.Run("should handle multiple tokens found", func(t *testing.T) {
		content := "workspace declaration"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 2) {
			assert.Equal(t, TokenKeyword, tokens[1].Type)
			assert.Equal(t, "declaration", tokens[1].Content)
		}
	})

	t.Run("should handle multiline", func(t *testing.T) {
		content := "workspace\ndeclaration"
		tokens, _ := Lexer(content)
		if assert.Equal(t, 3, len(tokens)) {
			assert.Equal(t, TokenNewline, tokens[1].Type)
			assert.Equal(t, "\n", tokens[1].Content)
		}
	})

	t.Run("should report token position", func(t *testing.T) {
		content := "workspace declaration"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 2) {
			assert.Equal(t, 0, tokens[1].Location.Line)
			assert.Equal(t, 10, tokens[1].Location.Pos)
		}
	})

	t.Run("should advance token position with multiple lines", func(t *testing.T) {
		content := "workspace\ndeclaration"
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 3) {
			assert.Equal(t, 1, tokens[2].Location.Line)
			assert.Equal(t, 0, tokens[2].Location.Pos)
		}
	})

	t.Run("should return string literals when found", func(t *testing.T) {
		content := `"name"`
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenString, tokens[0].Type)
			assert.Equal(t, "name", tokens[0].Content)
		}
	})
	t.Run("should return unterminated string literals when found", func(t *testing.T) {
		content := `"name`
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenString, tokens[0].Type)
			assert.Equal(t, "name", tokens[0].Content)
			assert.Equal(t, false, tokens[0].Terminated)
		}
	})
	t.Run("should return terminated string literals when found", func(t *testing.T) {
		content := `"name"`
		tokens, _ := Lexer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenString, tokens[0].Type)
			assert.Equal(t, "name", tokens[0].Content)
			assert.Equal(t, true, tokens[0].Terminated)
		}
	})
	t.Run("should handle escaped string literals when found", func(t *testing.T) {
		content := `"name with \"another string\""`
		tokens, _ := Lexer(content)
		if assert.Equal(t, 1, len(tokens)) {
			assert.Equal(t, TokenString, tokens[0].Type)
			assert.Equal(t, "name with \"another string\"", tokens[0].Content)
			assert.Equal(t, true, tokens[0].Terminated)
		}
	})
	t.Run("should handle equal sign when found", func(t *testing.T) {
		content := "identifier ="
		tokens, _ := Lexer(content)
		if assert.Equal(t, 2, len(tokens)) {
			assert.Equal(t, TokenEqual, tokens[1].Type)
			assert.Equal(t, "=", tokens[1].Content)
		}
	})
	t.Run("should handle equal sign when found", func(t *testing.T) {
		content := "identifier -> other"
		tokens, _ := Lexer(content)
		if assert.Equal(t, 3, len(tokens)) {
			assert.Equal(t, TokenRelation, tokens[1].Type)
		}
	})
	t.Run("should continue after symbols", func(t *testing.T) {
		content := "identifier = component"
		tokens, _ := Lexer(content)
		if assert.Equal(t, 3, len(tokens)) {
			assert.Equal(t, TokenKeyword, tokens[2].Type)
			assert.Equal(t, "component", tokens[2].Content)
		}
	})
	t.Run("one character keywords are handled properly", func(t *testing.T) {
		content := "a = b"
		tokens, _ := Lexer(content)
		assert.Equal(t, 3, len(tokens))
	})
	t.Run("exclamation mark still considered keyword", func(t *testing.T) {
		content := "!docs docs"
		tokens, _ := Lexer(content)
		assert.Equal(t, 2, len(tokens))
		assert.Equal(t, TokenKeyword, tokens[0].Type)
		assert.Equal(t, "!docs", tokens[0].Content)
	})
}
