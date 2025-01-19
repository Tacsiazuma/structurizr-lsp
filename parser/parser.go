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

func New(source string, content string) *Parser {
	tokens, _ := Lexer(source, content)
	return &Parser{tokens: tokens, root: NewNode(&Token{Content: "root"}, "root"), position: 0, diagnostics: make([]*Diagnostic, 0)}
}

type Workspace struct {
	model *Model
	views *ViewSet
}

type ASTNode struct {
	Token
	Type       string     // The type of the node (e.g., "workspace")
	Value      string     // The value of the node (if applicable, e.g., "name" or "description")
	Attributes []string   // Key-value pairs for additional attributes (e.g., {"name": "name", "description": "description"})
	Children   []*ASTNode // Nested nodes (e.g., body of the workspace)
}

func NewNode(token *Token, t string) *ASTNode {
	return &ASTNode{Token: *token, Type: t, Attributes: make([]string, 0), Children: make([]*ASTNode, 0)}
}
func displayTree(node *ASTNode, prefix string, isLast bool) {
	var connector string
	if isLast {
		connector = "└──"
	} else {
		connector = "├──"
	}
	fmt.Printf("%s%s %s (Attributes: %v)\n", prefix, connector, node.Type+" "+node.Value, node.Attributes)

	newPrefix := prefix
	if isLast {
		newPrefix += "    "
	} else {
		newPrefix += "│   "
	}

	for i, child := range node.Children {
		displayTree(child, newPrefix, i == len(node.Children)-1)
	}
}
func (n *ASTNode) ToString() string {
	attributes := mapToString(n.Attributes)
	children := ""
	if n.Children != nil {
		for _, c := range n.Children {
			children += c.ToString()
		}
	}
	return fmt.Sprintf("(%s %s %s)", n.Content, attributes, children)
}

func (n *ASTNode) AddChild(c *ASTNode) {
	n.Children = append(n.Children, c)
}

func (n *ASTNode) HasChild(t TokenType) bool {
	for _, c := range n.Children {
		if c.Token.Type == t {
			return true
		}
	}
	return false

}
func mapToString(m []string) string {
	var builder strings.Builder

	for _, value := range m {
		builder.WriteString(fmt.Sprintf("(%s) ", value))
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
	p.parse(p.root)
	displayTree(p.root, "", false)
	return p.root, p.diagnostics
}

func (p *Parser) parse(parent *ASTNode) {
	for {
		if !p.hasTokens() {
			return
		}
		tokens := p.readLine()
		var current *ASTNode
		for i, t := range tokens {
			switch t.Type {
			case TokenKeyword:
				// lookahead for equal if we are the first in the line
				if i+1 < len(tokens) && i == 0 && tokens[1].Type == TokenEqual {
					assign := NewNode(tokens[1], "assignment")
					parent.AddChild(assign)
					parent = assign
				}
				if current == nil {
					current = NewNode(t, string(t.Type))
					parent.AddChild(current)
				} else if parent.Token.Type == TokenEqual {
					current = NewNode(t, string(t.Type))
					parent.AddChild(current)
				} else {
					// handle subsequent keywords as attributes
					current.Attributes = append(current.Attributes, t.Content)
				}
			case TokenEqual:
				continue
			case TokenString:
				if current == nil {
					p.addDiagnostic(DiagnosticError, "Unexpected token, expected keyword got string", t.Location)
					return
				}
				current.Attributes = append(current.Attributes, t.Content)
			case TokenBraceOpen:
				if i+1 != len(tokens) {
					p.addDiagnostic(DiagnosticError, "Opening curly brace symbols ({) must be on the same line.", t.Location)
					return
				}
				p.parse(current)
				// one level down
			case TokenBraceClose:
				if i != 0 {
					p.addDiagnostic(DiagnosticError, "Closing curly brace symbols (}) must be on a line of their own.", t.Location)
				}
				return
			}
		}
		// return if we ended the end
	}
}

func (p *Parser) readLine() []*Token {
	tokens := make([]*Token, 0)
	for {
		t := p.nextToken()
		if t.Type != TokenEof && t.Type != TokenNewline {
			tokens = append(tokens, t)
		} else {
			return tokens
		}
	}
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

func (p *Parser) hasTokens() bool {
	return p.position < len(p.tokens)
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

func (p *Parser) addOptionalAttributes(node *ASTNode, t TokenType) {
	token := p.match(t)
	if token != nil {
		node.Attributes = append(node.Attributes, token.Content)
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

func (p *Parser) expectSequence(types ...TokenType) bool {
	for _, t := range types {
		if p.expect(t) == nil {
			return false
		}
	}
	return true
}

func (p *Parser) addDiagnostic(severity DiagnosticSeverity, message string, location Location) {
	fmt.Printf("%s %s %d:%d", severity, message, location.Line, location.Pos)
	p.diagnostics = append(p.diagnostics, &Diagnostic{
		Severity: severity,
		Message:  message,
		Location: location,
	})
}

// // line by line
// // keyword first
// // strings added as attributes
// // { goes one level deeper to add children
// func (p *Parser) consumeWorkspace() *ASTNode {
// 	p.consume(TokenComment, TokenNewline)
// 	if workspace := p.expect(TokenWorkspace); workspace == nil {
// 		return nil
// 	} else {
// 		p.root = NewNode(workspace, "workspace")
// 	}
// 	p.addOptionalAttributes(p.root, TokenString)
// 	p.addOptionalAttributes(p.root, TokenString)
// 	if !p.expectSequence(TokenBraceOpen, TokenNewline) {
// 		return nil
// 	}
// loop:
// 	for {
// 		t := p.peek()
// 		switch t.Type {
// 		case TokenModel:
// 			p.root.AddChild(p.consumeModel(t))
// 		case TokenViews:
// 			p.root.AddChild(p.consumeViews(t))
// 		case TokenNewline:
// 			p.consume(TokenNewline)
// 		case TokenBraceClose:
// 			break loop
// 		case TokenKeyword:
// 			p.root.AddChild(p.consumeKeyword(t))
// 		case TokenEof:
// 			break loop
// 		default:
// 			break loop
// 		}
// 	}
// 	if !p.root.HasChild(TokenModel) || !p.root.HasChild(TokenViews) {
// 		p.addDiagnostic(DiagnosticError, "Workspace must contain model and views", p.root.Location)
// 	}
// 	return p.root
// }
//
// func (p *Parser) consumeKeyword(t *Token) *ASTNode {
//
// }
//
// func (p *Parser) consumeModel(t *Token) *ASTNode {
// 	node := NewNode(t, "model")
// 	if !p.expectSequence(TokenModel, TokenBraceOpen) {
// 		return nil
// 	}
// loop:
// 	for {
// 		t := p.peek()
// 		switch t.Type {
// 		case TokenNewline:
// 			p.consume(TokenNewline)
// 		case TokenBraceClose:
// 			p.consume(TokenBraceClose)
// 			return node
// 		default:
// 			break loop
// 		}
// 	}
// 	return node
// }
//
// func (p *Parser) consumeViews(t *Token) *ASTNode {
// 	node := NewNode(t, "views")
// 	if !p.expectSequence(TokenViews, TokenBraceOpen) {
// 		return nil
// 	}
// loop:
// 	for {
// 		t := p.peek()
// 		switch t.Type {
// 		case TokenNewline:
// 			p.consume(TokenNewline)
// 		case TokenBraceClose:
// 			p.consume(TokenBraceClose)
// 			return node
// 		default:
// 			break loop
// 		}
// 	}
// 	return node
// }
