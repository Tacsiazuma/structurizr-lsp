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
	if node.Token.Content == "root" {
		// required children workspace
		if len(node.Children) == 0 {
			s.diagnostics = append(s.diagnostics, &Diagnostic{Message: "File must contain a workspace", Severity: DiagnosticWarning, Location: node.Location})
			return
		}
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
