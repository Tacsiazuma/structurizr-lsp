package parser

import (
	"bufio"
	"log"
	"os"
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
	Source string
	Line   int
	Pos    int
}
type TokenType string

const (
	TokenKeyword    TokenType = "keyword"
	TokenString     TokenType = "string"
	TokenNewline    TokenType = "newline"
	TokenBraceOpen  TokenType = "{"
	TokenBraceClose TokenType = "}"
	TokenEqual      TokenType = "="
	TokenRelation   TokenType = "->"
	TokenComment    TokenType = "comment"
	TokenEof        TokenType = "EOF"
)

var logger *log.Logger

func initLogger() {
	logFile, err := os.OpenFile("/home/tacsiazuma/work/structurizr-lsp/parser.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}

func Lexer(source string, content string, includer Includer) ([]Token, error) {
	initLogger()
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
		logger.Println("Scanning:" + text)
		switch state {
		case "start":
			if text == "\"" {
				state = "string"
				token = &Token{Type: TokenString, Content: "", Location: Location{Source: source, Line: line, Pos: pos}}
			} else if text == "/" || text == "#" {
				state = "singlelinecomment"
				token = &Token{Type: TokenComment, Content: text, Location: Location{Source: source, Line: line, Pos: pos}}
			} else if !unicode.IsSpace(rune(text[0])) {
				state = "keyword"
				token = &Token{Type: TokenKeyword, Content: text, Location: Location{Source: source, Line: line, Pos: pos}}
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
			} else if text == `"` || text == "\n" {
				token.Terminated = true
				tokens = append(tokens, *token)
				token = nil
				state = "start"
				escaped = false
			} else {
				token.Content += text
			}
		case "singlelinecomment":
			if text == "\n" {
				tokens = append(tokens, *token)
				token = nil
				state = "start"
			} else {
				token.Content += text
			}
			if token != nil && len(token.Content) == 2 && token.Content == "/*" {
				state = "multilinecomment"
			}
		case "multilinecomment":
			token.Content += text
			if strings.HasSuffix(token.Content, "*/") {
				tokens = append(tokens, *token)
				token = nil
				state = "start"
			}
		}
		if text == "\n" && state != "multilinecomment" {
			token = &Token{Type: TokenNewline, Content: "", Location: Location{Source: source, Line: line, Pos: pos}}
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
	tokens, err := checkIncludedFiles(tokens, source, includer)
	tokens = append(tokens, Token{Type: TokenEof, Content: "EOF", Location: Location{Source: source, Line: line, Pos: pos}})
	return tokens, err
}

func checkIncludedFiles(tokens []Token, source string, in Includer) ([]Token, error) {
	result := make([]Token, 0)
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Content == "!include" && i+1 < len(tokens) {
			path := tokens[i+1].Content
			content, err := in.include(source, path)
			included, err := Lexer(path, content, in)
			if err != nil {
				return nil, err
			}
			result = append(result, tokens[i])
			result = append(result, tokens[i+1])
			result = append(result, included[:len(included)-1]...)
			i++
		} else {
			result = append(result, tokens[i])
		}
	}
	return result, nil
}

func categorize(token *Token) {
	switch token.Content {
	case "{":
		token.Type = TokenBraceOpen
	case "}":
		token.Type = TokenBraceClose
	case "=":
		token.Type = TokenEqual
	case "->":
		token.Type = TokenRelation
	}
}
