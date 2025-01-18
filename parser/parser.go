package parser

import (
	"fmt"
)

type Parser struct {
	tokens      []Token
	position    int
	diagnostics []*Diagnostic
}

func New(content string) *Parser {
	tokens, _ := Lexer(content)
	return &Parser{tokens: tokens, position: 0, diagnostics: make([]*Diagnostic, 0)}
}

type Workspace struct {
	model *Model
	views *ViewSet
}
type Model struct{}
type ViewSet struct{}

type DiagnosticSeverity string

var (
	DiagnosticError DiagnosticSeverity = "error"
)

type Diagnostic struct {
	Message  string
	Location Location
	Severity DiagnosticSeverity
}

func (p *Parser) Parse() (*Workspace, []*Diagnostic) {
    workspace := p.consumeWorkspace()
	return workspace, p.diagnostics
}

/**
* This function consumes until it hits the end of the token stream or finds anything which is different
 */
func (p *Parser) consume(types ...TokenType) {
	for {
		current := p.peek()
		match := false
		for _, v := range types {
			if v == current.Type {
				match = true
			}
		}
		if match {
			fmt.Println("Consuming ", current.Type)
			p.nextToken()
		} else {
			break
		}
	}
}

func (p *Parser) nextToken() *Token {
	next := &p.tokens[p.position]
	p.position++
	return next
}

func (p *Parser) peek() *Token {
	current := &p.tokens[p.position]
	return current
}

func (p *Parser) match(t TokenType) *Token {
	current := p.peek()
	if current.Type == t {
		p.nextToken()
		return current
	} else {
		return nil
	}
}

func (p *Parser) expect(t TokenType) *Token {
	current := p.peek()
	if current.Type == t {
		p.nextToken()
		return current
	} else {
		p.addDiagnostic(DiagnosticError, fmt.Sprintf("Expected %s but found %s", t, current.Content), current.Location)
		return nil
	}
}

func (p *Parser) addDiagnostic(severity DiagnosticSeverity, message string, location Location) {
	p.diagnostics = append(p.diagnostics, &Diagnostic{
		Severity: severity,
		Message:  message,
		Location: location,
	})
}

func (p *Parser) consumeWorkspace() *Workspace {
	p.consume(TokenComment, TokenNewline)
	if p.expect(TokenWorkspace) == nil {
		return nil
	}
	p.match(TokenString)
	p.match(TokenString)
	if p.expect(TokenBraceOpen) == nil {
		return nil
	}
	if p.expect(TokenNewline) == nil {
		return nil
	}
	model := p.consumeModel()
	views := p.consumeViews()
	// consumeCloseBracket()
	return &Workspace{model: model, views: views}
}

func (p *Parser) consumeModel() *Model {
	p.consume(TokenComment, TokenNewline)
	if p.expect(TokenModel) == nil {
		return nil
	}
	if p.expect(TokenBraceOpen) == nil {
		return nil
	}
	if p.expect(TokenNewline) == nil {
		return nil
	}
	if p.expect(TokenBraceClose) == nil {
		return nil
	}
	return &Model{}
}

func (p *Parser) consumeViews() *ViewSet {
	p.consume(TokenComment, TokenNewline)
	if p.expect(TokenViews) == nil {
		return nil
	}
	if p.expect(TokenBraceOpen) == nil {
		return nil
	}
	if p.expect(TokenNewline) == nil {
		return nil
	}
	if p.expect(TokenBraceClose) == nil {
		return nil
	}
	return &ViewSet{}
}
