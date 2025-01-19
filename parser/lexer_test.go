package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	fake := &FakeIncluder{}
	file := "first.dsl"
	t.Run("should return EOF tokens on empty file", func(t *testing.T) {
		content := ""
		tokens, err := Lexer(file, content, fake)
		if assert.Nil(t, err) {
			assert.Equal(t, TokenEof, tokens[0].Type)
		}
	})
	t.Run("should return open brace symbol when found", func(t *testing.T) {
		content := "{"
		tokens, _ := Lexer(file, content, fake)
		assert.Equal(t, TokenBraceOpen, tokens[0].Type)
	})
	t.Run("should return close brace symbol when found", func(t *testing.T) {
		content := "}"
		tokens, _ := Lexer(file, content, fake)
		assert.Equal(t, TokenBraceClose, tokens[0].Type)
	})
	t.Run("should handle multiple tokens found", func(t *testing.T) {
		content := "workspace declaration"
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, len(tokens), 3) {
			assert.Equal(t, TokenKeyword, tokens[1].Type)
			assert.Equal(t, "declaration", tokens[1].Content)
		}
	})

	t.Run("should handle multiline", func(t *testing.T) {
		content := "workspace\ndeclaration"
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, 4, len(tokens)) {
			assert.Equal(t, TokenNewline, tokens[1].Type)
			assert.Equal(t, "", tokens[1].Content)
		}
	})

	t.Run("should report token position", func(t *testing.T) {
		content := "workspace declaration"
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, len(tokens), 3) {
			assert.Equal(t, 0, tokens[1].Location.Line)
			assert.Equal(t, 10, tokens[1].Location.Pos)
		}
	})

	t.Run("should advance token position with multiple lines", func(t *testing.T) {
		content := "workspace\ndeclaration"
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, len(tokens), 4) {
			assert.Equal(t, 1, tokens[2].Location.Line)
			assert.Equal(t, 0, tokens[2].Location.Pos)
		}
	})

	t.Run("should return string literals when found", func(t *testing.T) {
		content := `"name"`
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, len(tokens), 2) {
			assert.Equal(t, TokenString, tokens[0].Type)
			assert.Equal(t, "name", tokens[0].Content)
		}
	})
	t.Run("should return unterminated string literals when found", func(t *testing.T) {
		content := `"name`
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, len(tokens), 2) {
			assert.Equal(t, TokenString, tokens[0].Type)
			assert.Equal(t, "name", tokens[0].Content)
			assert.Equal(t, false, tokens[0].Terminated)
		}
	})
	t.Run("should return terminated string literals when found", func(t *testing.T) {
		content := `"name"`
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, len(tokens), 2) {
			assert.Equal(t, TokenString, tokens[0].Type)
			assert.Equal(t, "name", tokens[0].Content)
			assert.Equal(t, true, tokens[0].Terminated)
		}
	})
	t.Run("should handle escaped string literals when found", func(t *testing.T) {
		content := `"name with \"another string\""`
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, 2, len(tokens)) {
			assert.Equal(t, TokenString, tokens[0].Type)
			assert.Equal(t, "name with \"another string\"", tokens[0].Content)
			assert.Equal(t, true, tokens[0].Terminated)
		}
	})
	t.Run("should handle equal sign when found", func(t *testing.T) {
		content := "identifier ="
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, 3, len(tokens)) {
			assert.Equal(t, TokenEqual, tokens[1].Type)
			assert.Equal(t, "=", tokens[1].Content)
		}
	})
	t.Run("should handle included tokens when found after a newline", func(t *testing.T) {
		content := "!include test.dsl"
		tokens, _ := Lexer(file, content, fake)
		assert.Equal(t, "!include", tokens[0].Content)
		assert.Equal(t, "test.dsl", tokens[1].Content)
		assert.Equal(t, TokenNewline, tokens[2].Type)
		assert.Equal(t, "user", tokens[3].Content)
		assert.Equal(t, "Person", tokens[4].Content)
		assert.Equal(t, TokenEof, tokens[5].Type)
	})
	t.Run("included tokens are located in different file", func(t *testing.T) {
		content := "!include test.dsl"
		tokens, _ := Lexer(file, content, fake)
		assert.Equal(t, "first.dsl", tokens[0].Location.Source)
		assert.Equal(t, "first.dsl", tokens[1].Location.Source)
		assert.Equal(t, "first.dsl", tokens[2].Location.Source)
		assert.Equal(t, "test.dsl", tokens[3].Location.Source)
		assert.Equal(t, "test.dsl", tokens[4].Location.Source)
		assert.Equal(t, "first.dsl", tokens[5].Location.Source)
	})
	t.Run("should use absolute path for the included location", func(t *testing.T) {
		path, _ := os.Getwd()
		file := filepath.Join(path, "first.dsl")
		content := "!include test.dsl"
		tokens, _ := Lexer(file, content, fake)
		assert.Equal(t, filepath.Join(path, "first.dsl"), tokens[1].Location.Source)
		assert.Equal(t, filepath.Join(path, "test.dsl"), tokens[3].Location.Source)
		assert.Equal(t, filepath.Join(path, "test.dsl"), tokens[4].Location.Source)
	})
	t.Run("should handle relation sign when found", func(t *testing.T) {
		content := "identifier -> other"
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, 4, len(tokens)) {
			assert.Equal(t, TokenRelation, tokens[1].Type)
		}
	})
	t.Run("should handle multiple newlines alone", func(t *testing.T) {
		content := "workspace \" \n{\n"
		tokens, _ := Lexer(file, content, fake)
		if assert.Equal(t, 6, len(tokens)) {
			assert.Equal(t, TokenString, tokens[1].Type)
		}
	})
	t.Run("one character keywords are handled properly", func(t *testing.T) {
		content := "a = b"
		tokens, _ := Lexer(file, content, fake)
		assert.Equal(t, 4, len(tokens))
	})
}
