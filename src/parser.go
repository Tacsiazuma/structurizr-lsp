package tokenizer

import (
	"bufio"
	"strings"
	"unicode"
)

type Token struct {
	Type       TokenType
	Content    string
	Location   Location
	Terminated bool
}

type Location struct {
	Line int
	Pos  int
}
type TokenType string

const (
	TokenKeyword TokenType = "keyword"
	TokenString  TokenType = "string"
)

func Tokenizer(content string) ([]Token, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Split(bufio.ScanRunes)
	tokens := make([]Token, 0)
	var token *Token
	state := "start"
	line := 0
	pos := 0
	escaped := false
	for scanner.Scan() {
		text := scanner.Text()
		switch state {
		case "start":
			if unicode.IsLetter(rune(text[0])) {
				state = "keyword"
				token = &Token{Type: TokenKeyword, Content: text, Location: Location{Line: line, Pos: pos}}
			}
			if text == "\"" {
				state = "string"
				token = &Token{Type: TokenString, Content: "", Location: Location{Line: line, Pos: pos}}
			}
		case "keyword":
			if unicode.IsLetter([]rune(text)[0]) {
				token.Content += text
			} else {
				tokens = append(tokens, *token)
				token = nil
				state = "start"
			}
		case "string":
			if escaped {
				escaped = false
				token.Content += text
			} else if text == "\\" {
				escaped = true
			} else if text == `"` {
				token.Terminated = true
				tokens = append(tokens, *token)
				token = nil
				state = "start"
				escaped = false
			} else {
				token.Content += text
			}
		}
		if text == "\n" {
			pos = 0
			line++
		} else {
			pos++
		}
	}
	if token != nil {
		tokens = append(tokens, *token)
	}
	return tokens, nil
}
