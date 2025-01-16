package parser

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
	TokenKeyword        TokenType = "keyword"
	TokenString         TokenType = "string"
	TokenNewline        TokenType = "newline"
	TokenWorkspace      TokenType = "workspace"
	TokenModel          TokenType = "model"
	TokenGroup          TokenType = "group"
	TokenBraceOpen      TokenType = "braceopen"
	TokenBraceClose     TokenType = "braceclose"
	TokenEqual          TokenType = "equal"
	TokenRelation       TokenType = "relation"
	TokenViews          TokenType = "views"
	TokenPerson         TokenType = "person"
	TokenContainer      TokenType = "container"
	TokenComponent      TokenType = "component"
	TokenComment        TokenType = "comment"
	TokenSoftwareSystem TokenType = "softwareSystem"
)

func Lexer(content string) ([]Token, error) {
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
			if text == "\"" {
				state = "string"
				token = &Token{Type: TokenString, Content: "", Location: Location{Line: line, Pos: pos}}
			} else if !unicode.IsSpace(rune(text[0])) {
				state = "keyword"
				token = &Token{Type: TokenKeyword, Content: text, Location: Location{Line: line, Pos: pos}}
			}
		case "keyword":
			if !unicode.IsSpace([]rune(text)[0]) {
				token.Content += text
			} else {
				categorize(token)
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
			token = &Token{Type: TokenNewline, Content: "", Location: Location{Line: line, Pos: pos}}
			tokens = append(tokens, *token)
			token = nil
			pos = 0
			line++
		} else {
			pos++
		}
	}
	if token != nil {
		categorize(token)
		tokens = append(tokens, *token)
	}
	return tokens, nil
}

func categorize(token *Token) {
	switch token.Content {
	case "workspace":
		token.Type = TokenWorkspace
	case "model":
		token.Type = TokenModel
	case "group":
		token.Type = TokenGroup
	case "{":
		token.Type = TokenBraceOpen
	case "}":
		token.Type = TokenBraceClose
	case "=":
		token.Type = TokenEqual
	case "->":
		token.Type = TokenRelation
	case "views":
		token.Type = TokenViews
	case "person":
		token.Type = TokenPerson
	case "container":
		token.Type = TokenContainer
	case "softwareSystem":
		token.Type = TokenSoftwareSystem
	case "component":
		token.Type = TokenComponent
	}
}
