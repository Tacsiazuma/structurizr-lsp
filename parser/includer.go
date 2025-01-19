package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Includer struct {
}

func NewIncluder() *Includer {
	return &Includer{}
}

func (f *Includer) include(current, included string) (string, error) {
	// Resolve the absolute path
	baseDir := filepath.Dir(current)
	absolutePath := filepath.Join(baseDir, included)
	// Get file info for the given path
	info, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat path: %w", err)
	}

	// Check if the path is a directory
	if info.IsDir() {
		return readDir(absolutePath)
	} else {
		return readFile(absolutePath)
	}
}

func readDir(path string) (string, error) {
	// Read all files in the directory
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}
	var sb strings.Builder
	// Iterate over files and process .dsl files
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		// Check file extension
		if strings.HasSuffix(entry.Name(), ".dsl") {
			fullPath := filepath.Join(path, entry.Name())

			// Open and read the .dsl file
			content, err := readFile(fullPath)
			if err != nil {
				return "", fmt.Errorf("failed to read file: %w", err)
			}
			sb.WriteString(content)
		}
	}
	return sb.String(), nil
}

func readFile(path string) (string, error) {

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file's contents
	contents, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file contents: %w", err)
	}

	return string(contents), nil
}
