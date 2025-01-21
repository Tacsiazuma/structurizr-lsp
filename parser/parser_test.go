package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	file := "test.dsl"
	fake := &FakeIncluder{}
	t.Run("expect keyword as a start", func(t *testing.T) {
		sut := New(file, "something", fake)
		_, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
	})
	t.Run("jump over comments", func(t *testing.T) {
		sut := New(file, "# comment\nsomething", fake)
		_, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
	})
	t.Run("opening braces must be closed", func(t *testing.T) {
		sut := New(file, "workspace {", fake)
		_, diagnostics := sut.Parse()
		assert.Equal(t, 1, len(diagnostics))
		assert.Equal(t, "Unexpected EOF, expected }", diagnostics[0].Message)
	})
	t.Run("closed braces not report diagnostics", func(t *testing.T) {
		sut := New(file, "workspace {\n}", fake)
		_, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
	})
	t.Run("workspace }", func(t *testing.T) {
		sut := New(file, "workspace \n}", fake)
		_, diagnostics := sut.Parse()
		assert.Equal(t, 1, len(diagnostics))
		assert.Equal(t, "Expected EOF, got }", diagnostics[0].Message)
	})
	t.Run("assignments are handled", func(t *testing.T) {
		sut := New(file, "a = workspace", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
		assert.Equal(t, "(root  (=  (a  )(workspace  )))", ast.ToString())
	})
	t.Run("string properties are not reported as error", func(t *testing.T) {
		sut := New(file, "\"key\" \"value\"", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
		assert.Equal(t, "(root  (key (value) ))", ast.ToString())
	})
	t.Run("nested assignments are handled", func(t *testing.T) {
		sut := New(file, "a = workspace {\n b = component\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
		assert.Equal(t, "(root  (=  (a  )(workspace  ({  )(=  (b  )(component  ))(}  ))))", ast.ToString())
	})
	t.Run("assignments with attributes are handled", func(t *testing.T) {
		sut := New(file, "a = workspace \"test\" {\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
		assert.Equal(t, "(root  (=  (a  )(workspace (test) ({  )(}  ))))", ast.ToString())
	})
	t.Run("includes files to the token stream", func(t *testing.T) {
		sut := New(file, "!include file.dsl", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, 0, len(diagnostics))
		assert.Equal(t, "(root  (!include (file.dsl) )(=  (a  )(workspace (test) )))", ast.ToString())
	})

	t.Run("workspace elements can be defined in any order", func(t *testing.T) {
		sut := New(file, "workspace \"name\" \"description\" {\n views {\n}\nmodel {\n}\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, "(root  (workspace (name) (description) ({  )(views  ({  )(}  ))(model  ({  )(}  ))(}  )))", ast.ToString())
		assert.Empty(t, diagnostics)
	})
	t.Run("second keyword in a line appear as attribute", func(t *testing.T) {
		sut := New(file, "workspace \"name\" \"description\" {\n !identifiers flat \n views {\n}\nmodel {\n}\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, "(root  (workspace (name) (description) ({  )(!identifiers (flat) )(views  ({  )(}  ))(model  ({  )(}  ))(}  )))", ast.ToString())
		assert.Empty(t, diagnostics)
	})
	t.Run("multiple level of children allowed", func(t *testing.T) {
		sut := New(file, "workspace {\n model {\nsystemContext \"context\"{\n container \"container\"{\ncomponent \"component\"\n}\n}\n}\n}", fake)
		ast, diagnostics := sut.Parse()
		assert.Equal(t, "(root  (workspace  ({  )(model  ({  )(systemContext (context) ({  )(container (container) ({  )(component (component) )(}  ))(}  ))(}  ))(}  )))", ast.ToString())
		assert.Empty(t, diagnostics)
	})

}
