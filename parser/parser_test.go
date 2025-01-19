package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	file := "test.dsl"
    fake := &FakeIncluder{}
	t.Run("expect workspace as a start", func(t *testing.T) {
		sut := New(file, "something", fake)
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
		sut := New(file, "# comment\nsomething", fake)
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
		sut := New(file, "workspace something", fake)
		_, diagnostics := sut.Parse()
		assert.Equal(t, "Expected { but found something", diagnostics[0].Message)
		assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
	})
	t.Run("workspace can be followed by string", func(t *testing.T) {
		sut := New(file, "workspace \"name\" {", fake)
		_, diagnostics := sut.Parse()
		assert.Equal(t, "Expected newline but found EOF", diagnostics[0].Message)
		assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
	})
	t.Run("workspace must contain model and views", func(t *testing.T) {
		sut := New(file, "workspace \"name\" {\n}", fake)
		_, diagnostics := sut.Parse()
		assert.Equal(t, "Workspace must contain model and views", diagnostics[0].Message)
		assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
	})

	t.Run("workspace must contain views", func(t *testing.T) {
		sut := New(file, "workspace \"name\" {\n model {\n}\n}", fake)
		_, diagnostics := sut.Parse()
		if assert.Equal(t, 1, len(diagnostics)) {
			assert.Equal(t, "Workspace must contain model and views", diagnostics[0].Message)
			assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
		}
	})

	t.Run("minimal workspace parses without errors", func(t *testing.T) {
		sut := New(file, "workspace \"name\" {\n model {\n}\nviews {\n}\n}", fake)
		_, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
	})
	t.Run("assignments are handled", func(t *testing.T) {
		sut := New(file, "a = workspace", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
		assert.Equal(t, "(root  (=  (a  )(workspace  )))", ast.ToString())
	})
	t.Run("nested assignments are handled", func(t *testing.T) {
		sut := New(file, "a = workspace {\n b = component\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
		assert.Equal(t, "(root  (=  (a  )(workspace  (=  (b  )(component  )))))", ast.ToString())
	})
	t.Run("assignments with attributes are handled", func(t *testing.T) {
		sut := New(file, "a = workspace \"test\" {\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
		assert.Equal(t, "(root  (=  (a  )(workspace (test) )))", ast.ToString())
	})
	t.Run("includes files to the token stream", func(t *testing.T) {
		sut := New(file, "!include file.dsl", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
		assert.Equal(t, "(root  (=  (a  )(workspace (test) )))", ast.ToString())
	})

	t.Run("workspace elements can be defined in any order", func(t *testing.T) {
		sut := New(file, "workspace \"name\" \"description\" {\n views {\n}\nmodel {\n}\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, "(root  (workspace (name) (description) (views  )(model  )))", ast.ToString())
		assert.Empty(t, diagnostics)
	})
	t.Run("second keyword in a line appear as attribute", func(t *testing.T) {
		sut := New(file, "workspace \"name\" \"description\" {\n !identifiers flat \n views {\n}\nmodel {\n}\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, "(root  (workspace (name) (description) (!identifiers (flat) )(views  )(model  )))", ast.ToString())
		assert.Empty(t, diagnostics)
	})
	t.Run("multiple level of children allowed", func(t *testing.T) {
		sut := New(file, "workspace {\n model {\nsystemContext \"context\"{\n container \"container\"{\ncomponent \"component\"\n}\n}\n}\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, "(root  (workspace  (model  (systemContext (context) (container (container) (component (component) ))))))", ast.ToString())
		assert.Empty(t, diagnostics)
	})
	t.Run("docs are added as a child", func(t *testing.T) {
		sut := New(file, "workspace \"name\" \"description\" {\n!docs docs\n views {\n}\nmodel {\n}\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, "(workspace (name) (description) (!docs docs) (views  )(model  ))", ast.ToString())
		assert.Empty(t, diagnostics)
	})

	t.Run("fails when too many strings for workspace", func(t *testing.T) {
		sut := New(file, "workspace \"name\" \"Description\" \"some\" {\n}", fake)
		_, diagnostics := sut.Parse()
		if assert.Equal(t, 1, len(diagnostics)) {
			assert.Equal(t, "Expected { but found some", diagnostics[0].Message)
			assert.Equal(t, DiagnosticError, diagnostics[0].Severity)
		}
	})
}
