package parser

import "sync"

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
		if c.Token.Content == "model" {
			s.visitModel(c)
		}
		if c.Token.Content == "views" {
			s.visitViews(c)
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

func (s *SemanticAnalyser) visitProperties(node *ASTNode) {
	logger.Println("visitProperties")
	for _, c := range node.Children {
		if c.Token.Type == TokenString {
			c.Token.Type = TokenName
			if len(c.Attributes) > 0 && c.Attributes[0].Type == TokenString {
				c.Attributes[0].Type = TokenValue
			}
		}
	}
}

func (s *SemanticAnalyser) visitModel(node *ASTNode) {
	logger.Println("visitModel")
	for _, c := range node.Children {
		if c.Token.Content == "person" {
			s.visitPerson(c)
		}
	}
	s.ws.model = &Model{}
}

func (s *SemanticAnalyser) visitPerson(node *ASTNode) {
	AugmentAttributes(node)
	logger.Println("visitPerson")
}
