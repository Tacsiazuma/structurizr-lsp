package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncluder(t *testing.T) {
	i := NewIncluder()
	t.Run("it should return file content in the same directory", func(t *testing.T) {
		content, err := i.include("includer_test.go", "included.dsl")
		assert.NoError(t, err)
		assert.Equal(t, "person \"User\"\n", content)
	})
	t.Run("it should work with absolute source path", func(t *testing.T) {
		cwd, _ := os.Getwd()
		content, err := i.include(filepath.Join(cwd, "includer_test.go"), "included.dsl")
		assert.NoError(t, err)
		assert.Equal(t, "person \"User\"\n", content)
	})

	t.Run("it should work with directories", func(t *testing.T) {
		cwd, _ := os.Getwd()
		content, err := i.include(filepath.Join(cwd, "includer_test.go"), "included")
		assert.NoError(t, err)
		assert.Equal(t, "first\nlast\n", content)
	})
	t.Run("it should return error when failing to read", func(t *testing.T) {
		cwd, _ := os.Getwd()
		_, err := i.include(filepath.Join(cwd, "includer_test.go"), "nonexistent")
		assert.Error(t, err)
	})
}
