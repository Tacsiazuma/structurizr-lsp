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
	if s.ws.Model == nil {
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
	model := &Model{
		People:                 make(map[string]*Person),
		Groups:                 make(map[string]*Group),
		References:             make(map[string]interface{}),
		SoftwareSystems:        make(map[string]*SoftwareSystem),
		DeploymentEnvironments: make(map[string]*DeploymentEnvironment),
	}
	for _, c := range node.Children {
		if c.Token.Content == "person" {
			person := s.visitPerson(c)
			model.People[person.Name] = person
		} else if isKeyWordWithName(c, "!identifiers") {
			model.Identifiers = s.visitOptionWithPossibleValues(c, "flat", "hierarchical")
		} else if isAssignment(c, "person") {
			person := s.visitPerson(c.Children[1])
			identifier := getIdentifier(c)
			model.References[identifier] = person
			model.People[person.Name] = person
		} else if isKeyWordWithName(c, "group") {
			model.Groups[fmt.Sprintf("%p", &c)] = s.visitGroup(c)
		} else if isKeyWordWithName(c, "softwareSystem") {
			ss := s.visitSoftwareSystem(c)
			model.SoftwareSystems[ss.Name] = ss
		} else if isKeyWordWithName(c, "deploymentEnvironment") {
			de := s.visitDeploymentEnvironment(c)
			model.DeploymentEnvironments[de.Name] = de
		}
	}
	s.ws.Model = model
}

func (s *SemanticAnalyser) visitGroup(node *ASTNode) *Group {
	AugmentAttributes(node)
	logger.Println("visitGroup")
	return &Group{Name: node.Attributes[0].Content}
}

func (s *SemanticAnalyser) visitSoftwareSystem(node *ASTNode) *SoftwareSystem {
	AugmentAttributes(node)
	logger.Println("visitSoftwareSystem")
	return &SoftwareSystem{Name: node.Attributes[0].Content}
}

func (s *SemanticAnalyser) visitDeploymentEnvironment(node *ASTNode) *DeploymentEnvironment {
	AugmentAttributes(node)
	logger.Println("visitDeploymentEnvironment")
	return &DeploymentEnvironment{Name: node.Attributes[0].Content}
}

func isAssignment(node *ASTNode, t string) bool {
	return node.Type == "assignment" && node.Children[1].Content == t
}

// Visits a person node
func (s *SemanticAnalyser) visitPerson(node *ASTNode) *Person {
	AugmentAttributes(node)
	logger.Println("visitPerson")
	return &Person{Name: node.Attributes[0].Content}
}

// Visits a person node
func (s *SemanticAnalyser) visitConfiguration(node *ASTNode) *Configuration {
	config := &Configuration{}
	for _, c := range node.Children {
		if isKeyWordWithName(c, "scope") {
			config.Scope = s.visitOptionWithPossibleValues(c, "landscape", "softwaresystem", "none")
		} else if isKeyWordWithName(c, "visibility") {
			config.Visibility = s.visitOptionWithPossibleValues(c, "private", "public")
		} else if isKeyWordWithName(c, "users") {
			config.Users = s.visitUsers(c)
		} else if isKeyWordWithName(c, "properties") {
			config.Properties = s.visitProperties(c)
		} else if isBraces(c) {
			continue
		} else {
			s.addWarning("Unexpected children: "+c.Token.Content, c)
		}
	}
	return config
}

func getIdentifier(c *ASTNode) string {
	return c.Children[0].Content
}

func (s *SemanticAnalyser) visitUsers(node *ASTNode) map[string]string {
	props := make(map[string]string)
	for _, c := range node.Children {
		if c.Token.Type == TokenKeyword {
			props[c.Token.Content] = s.visitOptionWithPossibleValues(c, "write", "read")
		}
	}
	return props
}
