package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	t.Run("expect workspace as a start", func(t *testing.T) {
		sut := New("something")
		workspace, diagnostics := sut.Parse()
		assert.Nil(t, workspace)
		assert.Equal(t, 1, len(diagnostics))
		d := diagnostics[0]
		assert.Equal(t, "Expected workspace but found something", d.Message)
		assert.Equal(t, 0, d.Location.Line)
		assert.Equal(t, 0, d.Location.Pos)
		assert.Equal(t, DiagnosticError, d.Severity)
	})
	t.Run("jump over comments", func(t *testing.T) {
		sut := New("# comment\nsomething")
		workspace, diagnostics := sut.Parse()
		assert.Nil(t, workspace)
		assert.Equal(t, 1, len(diagnostics))
		d := diagnostics[0]
		assert.Equal(t, "Expected workspace but found something", d.Message)
		assert.Equal(t, 1, d.Location.Line)
		assert.Equal(t, 0, d.Location.Pos)
		assert.Equal(t, DiagnosticError, d.Severity)
	})
	t.Run("workspace cannot be followed by keyword", func(t *testing.T) {
		sut := New("workspace something")
		workspace, diagnostics := sut.Parse()
		assert.Nil(t, workspace)
		assert.Equal(t, "Expected { but found something", diagnostics[0].Message)
		assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
	})
	t.Run("workspace can be followed by string", func(t *testing.T) {
		sut := New("workspace \"name\" {")
		workspace, diagnostics := sut.Parse()
		assert.Nil(t, workspace)
		assert.Equal(t, "Expected newline but found EOF", diagnostics[0].Message)
		assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
	})
	t.Run("workspace must contain model", func(t *testing.T) {
		sut := New("workspace \"name\" {\n}")
		workspace, diagnostics := sut.Parse()
		assert.Nil(t, workspace)
		assert.Equal(t, "Expected model but found }", diagnostics[0].Message)
		assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
	})

	t.Run("workspace must contain views", func(t *testing.T) {
		sut := New("workspace \"name\" {\n model {\n}\n}")
		workspace, diagnostics := sut.Parse()
		assert.Nil(t, workspace)
		if assert.Equal(t, 1, len(diagnostics)) {
			assert.Equal(t, "Expected views but found }", diagnostics[0].Message)
			assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
		}
	})

	t.Run("minimal workspace parses without errors", func(t *testing.T) {
		sut := New("workspace \"name\" {\n model {\n}\nviews {\n}\n}")
		_, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
	})

	t.Run("minimal workspace returns workspace", func(t *testing.T) {
		sut := New("workspace \"name\" {\n model {\n}\nviews {\n}\n}")
		workspace, _ := sut.Parse()
		assert.Equal(t, &Workspace{model: &Model{}, views: &ViewSet{}}, workspace)
	})
	t.Run("workspace elements can be defined in any order", func(t *testing.T) {
		sut := New("workspace \"name\" {\n views {\n}\nmodel {\n}\n}")
		workspace, diagnostics := sut.Parse()
        assert.Empty(t, diagnostics)
		assert.Equal(t, "(workspace (name name) (description nil))", workspace.ToString())
	})

	t.Run("fails when too many strings for workspace", func(t *testing.T) {
		sut := New("workspace \"name\" \"Description\" \"some\" {\n}")
		_, diagnostics := sut.Parse()
		if assert.Equal(t, 1, len(diagnostics)) {
			assert.Equal(t, "Expected { but found some", diagnostics[0].Message)
			assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
		}
	})
}
