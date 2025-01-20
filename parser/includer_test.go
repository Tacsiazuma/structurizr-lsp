package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncluder(t *testing.T) {
	i := NewIncluder()
	cwd, _ := os.Getwd()
	t.Run("it should return file content in the same directory", func(t *testing.T) {
		content, err := i.include(filepath.Join(cwd, "included.dsl"))
		assert.NoError(t, err)
		assert.Equal(t, "person \"User\"\n", content)
	})
	t.Run("it should work with absolute source path", func(t *testing.T) {
		content, err := i.include(filepath.Join(cwd, "included.dsl"))
		assert.NoError(t, err)
		assert.Equal(t, "person \"User\"\n", content)
	})

	t.Run("it should work with URI", func(t *testing.T) {
		content, err := i.include(filepath.Join(cwd, "included.dsl"))
		assert.NoError(t, err)
		assert.Equal(t, "person \"User\"\n", content)
	})
	t.Run("it should work with directories", func(t *testing.T) {
		content, err := i.include(filepath.Join(cwd, "included"))
		assert.NoError(t, err)
		assert.Equal(t, "first\nlast\n", content)
	})

	t.Run("it should return error when failing to read", func(t *testing.T) {
		_, err := i.include("nonexistent")
		assert.Error(t, err)
	})
}
