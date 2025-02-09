package parser

import (
	"fmt"
	"sync"
)

type SemanticAnalyser struct {
	parser      *Parser
	diagnostics []*Diagnostic
	ws          *Workspace
	mu          sync.Mutex
}

func (s *SemanticAnalyser) Analyse() (*Workspace, *ASTNode, []*Diagnostic) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ast, diag := s.parser.Parse()
	s.diagnostics = diag
	s.walk(ast)
	return s.ws, ast, s.diagnostics
}

func (s *SemanticAnalyser) walk(node *ASTNode) {
	if node == nil {
		return
	}
	if node.Token.Content == "root" {
		s.visitRoot(node)
	}
}

func NewAnalyser(sourceFile string, content string) *SemanticAnalyser {
	p := New(sourceFile, content, NewIncluder())
	return &SemanticAnalyser{parser: p}
}

func (s *SemanticAnalyser) visitRoot(node *ASTNode) {
	// required children workspace
	for _, c := range node.Children {
		if c.Token.Content == "workspace" {
			s.visitWorkspace(c)
		}
	}
	if s.ws == nil {
		s.addWarning("File must contain a workspace", node)
	}
}

func (s *SemanticAnalyser) visitWorkspace(node *ASTNode) {
	logger.Println("visitWorkspace")
	s.ws = &Workspace{}
	for _, c := range node.Children {
		if isKeyWordWithName(c, "model") {
			s.visitModel(c)
		} else if isKeyWordWithName(c, "views") {
			s.visitViews(c)
		} else if isKeyWordWithName(c, "name") {
			s.ws.Name = s.visitAttribute(c)
		} else if isKeyWordWithName(c, "properties") {
			s.ws.Properties = s.visitProperties(c)
		} else if isKeyWordWithName(c, "description") {
			s.ws.Description = s.visitAttribute(c)
		} else if isKeyWordWithName(c, "!identifiers") {
			s.ws.Identifiers = s.visitOptionWithPossibleValues(c, "flat", "hierarchical")
		} else if isKeyWordWithName(c, "!docs") {
			s.ws.Docs = s.visitDocs(c)
		} else if isKeyWordWithName(c, "!adrs") {
			s.ws.Adrs = s.visitAdrs(c)
		} else if isKeyWordWithName(c, "configuration") {
			s.ws.Configuration = s.visitConfiguration(c)
		} else if isBraces(c) {
			continue
		} else {
			s.addWarning("Unexpected children: "+c.Token.Content, c)
		}
	}
	AugmentAttributes(node)
	if s.ws.model == nil {
		s.addWarning("Workspace must contain a model", node)
	}
	if s.ws.views == nil {
		s.addWarning("Workspace must contain views", node)
	}
}

func isBraces(node *ASTNode) bool {
	return node.Token.Type == TokenBraceClose || TokenBraceOpen == node.Token.Type
}

func (s *SemanticAnalyser) visitDocs(node *ASTNode) *Documentation {
	docs := &Documentation{}
	if len(node.Attributes) > 0 && node.Attributes[0].Type == TokenKeyword {
		docs.Path = node.Attributes[0].Content
	}
	if len(node.Attributes) > 1 && node.Attributes[1].Type == TokenKeyword {
		docs.Fqcn = node.Attributes[1].Content
	}
	return docs
}

func (s *SemanticAnalyser) visitAdrs(node *ASTNode) *ADR {
	adrs := &ADR{}
	if len(node.Attributes) > 0 && node.Attributes[0].Type == TokenKeyword {
		adrs.Path = node.Attributes[0].Content
	}
	if len(node.Attributes) > 1 && node.Attributes[1].Type == TokenKeyword {
		adrs.Fqcn = node.Attributes[1].Content
	}
	return adrs
}

func isKeyWordWithName(node *ASTNode, name string) bool {
	if node.Token.Type == TokenKeyword && node.Token.Content == name {
		return true
	}
	return false
}

func (s *SemanticAnalyser) visitOptionWithPossibleValues(node *ASTNode, possibleValues ...string) string {
	if len(node.Attributes) > 0 && node.Attributes[0].Type == TokenKeyword {
		for _, v := range possibleValues {
			if v == node.Attributes[0].Content {
				return v
			}
		}
	}
	s.addWarning(fmt.Sprintf("Invalid option, possible values %s", possibleValues), node)
	return ""
}

func (s *SemanticAnalyser) visitAttribute(node *ASTNode) string {
	if len(node.Attributes) > 0 && node.Attributes[0].Type == TokenString {
		return node.Attributes[0].Content
	}
	return ""
}

func AugmentAttributes(node *ASTNode) {
	if len(node.Attributes) > 0 && node.Attributes[0].Type == TokenString {
		node.Attributes[0].Type = TokenName
	}
	if len(node.Attributes) > 1 && node.Attributes[1].Type == TokenString {
		node.Attributes[1].Type = TokenDescription
	}
	if len(node.Attributes) > 2 && node.Attributes[2].Type == TokenString {
		node.Attributes[2].Type = TokenTags
	}
}

func (s *SemanticAnalyser) addWarning(message string, node *ASTNode) {
	s.diagnostics = append(s.diagnostics, &Diagnostic{Message: message, Severity: DiagnosticWarning, Location: node.Location})
}

func (s *SemanticAnalyser) visitViews(node *ASTNode) {
	logger.Println("visitViews")
	for _, c := range node.Children {
		logger.Println(c.Token.Content)
		if c.Token.Content == "properties" && c.Token.Type == TokenKeyword {
			s.visitProperties(c)
		}
	}
	s.ws.views = &ViewSet{}
}

func (s *SemanticAnalyser) visitProperties(node *ASTNode) map[string]string {
	logger.Println("visitProperties")
	props := make(map[string]string)
	for _, c := range node.Children {
		if c.Token.Type == TokenString {
			c.Token.Type = TokenName
			if len(c.Attributes) > 0 && c.Attributes[0].Type == TokenString {
				c.Attributes[0].Type = TokenValue
				props[c.Token.Content] = c.Attributes[0].Content
			}
		}
	}
	return props
}

func (s *SemanticAnalyser) visitModel(node *ASTNode) {
	logger.Println("visitModel")
	s.ws.model = &Model{}
	for _, c := range node.Children {
		if c.Token.Content == "person" {
			s.visitPerson(c)
		}
	}
}

// Visits a person node
func (s *SemanticAnalyser) visitPerson(node *ASTNode) {
	AugmentAttributes(node)
	logger.Println("visitPerson")
}

// Visits a person node
func (s *SemanticAnalyser) visitConfiguration(node *ASTNode) *Configuration {
	return &Configuration{}
}
