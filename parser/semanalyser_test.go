package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemanticAnalyser(t *testing.T) {
	t.Run("workspace required under the root", func(t *testing.T) {
		sut := NewTestAnalyser("/* something */")
		workspace, _, diags := sut.Analyse()
		assert.Nil(t, workspace)
		if assert.Equal(t, 1, len(diags)) {
			assert.Equal(t, "File must contain a workspace", diags[0].Message)
		}
	})
	t.Run("model and views required under workspace", func(t *testing.T) {
		sut := NewTestAnalyser("workspace {\n}")
		_, _, diags := sut.Analyse()
		if assert.Equal(t, 2, len(diags)) {
			assert.Equal(t, "Workspace must contain a model", diags[0].Message)
			assert.Equal(t, "Workspace must contain views", diags[1].Message)
		}
	})
	t.Run("minimal workspace without errors", func(t *testing.T) {
		sut := NewTestAnalyser("workspace {\nmodel {\n}\nviews {\n}\n}")
		_, _, diags := sut.Analyse()
		assert.Equal(t, 0, len(diags))
	})
	t.Run("augments workspace attributes", func(t *testing.T) {
		sut := NewTestAnalyser("workspace \"name\" \"description\" {\nmodel {\n}\nviews {\n}\n}")
		_, ast, _ := sut.Analyse()
		workspace := ast.Children[0]
		assert.Equal(t, TokenName, workspace.Attributes[0].Type)
		assert.Equal(t, TokenDescription, workspace.Attributes[1].Type)
	})
	t.Run("augments properties", func(t *testing.T) {
		sut := NewTestAnalyser("workspace \"name\" \"description\" {\nmodel {\n}\nviews {\nproperties {\n\"key\" \"value\"\n}\n}\n}")
		_, ast, _ := sut.Analyse()
		ws := ast.Children[0]
		views := ws.Children[2]
		properties := views.Children[1]
        property := properties.Children[1]
		assert.Equal(t, TokenName, property.Token.Type)
		assert.Equal(t, TokenValue, property.Attributes[0].Type)
	})
	t.Run("augments person attributes", func(t *testing.T) {
		sut := NewTestAnalyser("workspace {\nmodel {\nperson \"name\" \"description\" \"tags\" \n}\nviews {\n}\n}")
		_, ast, _ := sut.Analyse()
		ws := ast.Children[0]
		model := ws.Children[1]
		person := model.Children[1]
		assert.Equal(t, TokenName, person.Attributes[0].Type)
		assert.Equal(t, TokenDescription, person.Attributes[1].Type)
		assert.Equal(t, TokenTags, person.Attributes[2].Type)
	})
}

func NewTestAnalyser(content string) *SemanticAnalyser {
	return &SemanticAnalyser{parser: New("test.dsl", content, &FakeIncluder{})}
}
