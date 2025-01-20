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
		sut := NewTestAnalyser("workspace")
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
}

func NewTestAnalyser(content string) *SemanticAnalyser {
	return &SemanticAnalyser{parser: New("test.dsl", content, &FakeIncluder{})}
}
