package parser

import (
	"fmt"
	"strings"
)

type Parser struct {
	root        *ASTNode
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

type ASTNode struct {
	Token
	Type       string            // The type of the node (e.g., "workspace")
	Value      string            // The value of the node (if applicable, e.g., "name" or "description")
	Attributes map[string]string // Key-value pairs for additional attributes (e.g., {"name": "name", "description": "description"})
	Children   []*ASTNode        // Nested nodes (e.g., body of the workspace)
}

func NewNode(token *Token, t string) *ASTNode {
	return &ASTNode{Token: *token, Type: t, Attributes: make(map[string]string)}
}

func (n *ASTNode) ToString() string {
	attributes := mapToString(n.Attributes)
	return fmt.Sprintf("(%s %s)", n.Type, attributes)
}

func mapToString(m map[string]string) string {
	var builder strings.Builder

	for key, value := range m {
		builder.WriteString(fmt.Sprintf("(%s %s) ", key, value))
	}

	// Trim the trailing space
	result := strings.TrimSpace(builder.String())
	return result
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

func (p *Parser) Parse() (*ASTNode, []*Diagnostic) {
	ast := p.consumeWorkspace()
	return ast, p.diagnostics
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

func (p *Parser) matchAttribute(node *ASTNode, name string, t TokenType) {
	token := p.match(t)
	if token != nil {
		node.Attributes[name] = token.Content
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

func (p *Parser) consumeWorkspace() *ASTNode {
	// comments, newlines ignore
	// workspace token, wrap in node
	p.consume(TokenComment, TokenNewline)
	if workspace := p.expect(TokenWorkspace); workspace == nil {
		return nil
	} else {
		p.root = NewNode(workspace, "workspace")
	}
	p.matchAttribute(p.root, "name", TokenString)
	p.matchAttribute(p.root, "description", TokenString)
	if p.expect(TokenBraceOpen) == nil {
		return nil
	}
	if p.expect(TokenNewline) == nil {
		return nil
	}
	_ = p.consumeModel()
	_ = p.consumeViews()
	if p.expect(TokenNewline) == nil {
		return nil
	}
	if p.expect(TokenBraceClose) == nil {
		return nil
	}
	return p.root
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
