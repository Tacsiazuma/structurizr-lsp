package parser

type SemanticAnalyser struct {
	parser      *Parser
	diagnostics []*Diagnostic
	ws          *Workspace
}

func (s *SemanticAnalyser) Analyse() (*Workspace, *ASTNode, []*Diagnostic) {
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

func contains(slice []TokenType, t TokenType) bool {
	for _, item := range slice {
		if item == t {
			return true
		}
	}
	return false
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
		s.diagnostics = append(s.diagnostics, &Diagnostic{Message: "File must contain a workspace", Severity: DiagnosticWarning, Location: node.Location})
		return
	}
}

func (s *SemanticAnalyser) visitWorkspace(node *ASTNode) {
	s.ws = &Workspace{}
	for _, c := range node.Children {
		if c.Token.Content == "model" {
			s.visitModel(c)
		}
		if c.Token.Content == "views" {
			s.visitViews(c)
		}
	}
	if s.ws.model == nil {
		s.diagnostics = append(s.diagnostics, &Diagnostic{Message: "Workspace must contain a model", Severity: DiagnosticWarning, Location: node.Location})
		return
	}
	if s.ws.views == nil {
		s.diagnostics = append(s.diagnostics, &Diagnostic{Message: "Workspace must contain views", Severity: DiagnosticWarning, Location: node.Location})
		return
	}
}

func (s *SemanticAnalyser) visitViews(c *ASTNode) {
	s.ws.views = &ViewSet{}
}

func (s *SemanticAnalyser) visitModel(node *ASTNode) {
	s.ws.model = &Model{}
}
