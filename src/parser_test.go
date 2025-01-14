package tokenizer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTokenizer(t *testing.T) {
	t.Run("should return empty tokens on empty file", func(t *testing.T) {
		content := ""
		tokens, err := Tokenizer(content)
		if assert.Nil(t, err) {
			assert.Equal(t, make([]Token, 0), tokens)
		}
	})

	t.Run("should return workspace keyword when found", func(t *testing.T) {
		content := "workspace"
		tokens, _ := Tokenizer(content)
		if assert.Equal(t, len(tokens), 1) {
			assert.Equal(t, TokenKeyword, tokens[0].Type)
			assert.Equal(t, "workspace", tokens[0].Content)
		}
	})

	t.Run("should handle multiple tokens found", func(t *testing.T) {
		content := "workspace declaration"
		tokens, _ := Tokenizer(content)
		if assert.Equal(t, len(tokens), 2) {
			assert.Equal(t, TokenKeyword, tokens[1].Type)
			assert.Equal(t, "declaration", tokens[1].Content)
		}
	})

	t.Run("should handle multiline", func(t *testing.T) {
		content := "workspace\ndeclaration"
		tokens, _ := Tokenizer(content)
		if assert.Equal(t, len(tokens), 2) {
			assert.Equal(t, TokenKeyword, tokens[1].Type)
			assert.Equal(t, "declaration", tokens[1].Content)
		}
	})

	t.Run("should report token position", func(t *testing.T) {
		content := "workspace declaration"
		tokens, _ := Tokenizer(content)
		if assert.Equal(t, len(tokens), 2) {
			assert.Equal(t, 0, tokens[1].Location.Line)
			assert.Equal(t, 10, tokens[1].Location.Pos)
		}
	})

	t.Run("should report token position with multiple lines", func(t *testing.T) {
		content := "workspace\ndeclaration"
		tokens, _ := Tokenizer(content)
		if assert.Equal(t, len(tokens), 2) {
			assert.Equal(t, 1, tokens[1].Location.Line)
			assert.Equal(t, 0, tokens[1].Location.Pos)
		}
	})
}
