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
	t.Run("model required under workspace", func(t *testing.T) {
		sut := NewTestAnalyser("workspace")
		workspace, _, diags := sut.Analyse()
		assert.Nil(t, workspace)
		if assert.Equal(t, 1, len(diags)) {
			assert.Equal(t, "workspace must contain a model", diags[0].Message)
		}
	})
}

func NewTestAnalyser(content string) *SemanticAnalyser {
	return &SemanticAnalyser{parser: New("test.dsl", content, &FakeIncluder{})}
}
