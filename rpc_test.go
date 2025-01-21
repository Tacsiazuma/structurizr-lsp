package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCase holds the input and expected output for a test case.
type TestCase struct {
	Input  string
	Output string
}

func TestRpc(t *testing.T) {
	t.Run("initialize", func(t *testing.T) {
		writer := &UnbufferedWriter{}
		reader := &StringReader{}
		sut := From(reader, writer)
		testcase, _ := ParseTestFile("initialize.txt")
		reader.SetString(testcase.Input)
		sut.Handle()
		assert.Equal(t, testcase.Output, writer.written)
	})
}

// UnbufferedWriter writes data directly to an underlying destination.
type UnbufferedWriter struct {
	written string
}

// StringReader is a custom reader that allows setting a string later.
type StringReader struct {
	reader *strings.Reader
}

// SetString sets the string for the reader.
func (sr *StringReader) SetString(s string) {
	sr.reader = strings.NewReader(s)
}

// Read reads data from the string set earlier.
// It implements the io.Reader interface.
func (sr *StringReader) Read(p []byte) (int, error) {
	if sr.reader == nil {
		return 0, errors.New("no string provided")
	}
	return sr.reader.Read(p)
}

// Write implements the io.Writer interface.
func (w *UnbufferedWriter) Write(p []byte) (int, error) {
	w.written = string(p)
	return len(p), nil
}

// ParseTestFile reads and parses a test file into a TestCase.
func ParseTestFile(filepath string) (*TestCase, error) {
	// Read the file contents.
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Convert to a string and split by the markers.
	content := string(data)
	parts := strings.Split(content, "=== OUTPUT")
	if len(parts) != 2 {
		return nil, errors.New("test file must contain '=== OUTPUT' marker")
	}

	inputSection := strings.TrimSpace(strings.Split(parts[0], "=== INPUT")[1])
	outputSection := strings.TrimSpace(parts[1])

	return &TestCase{
		Input:  inputSection,
		Output: outputSection,
	}, nil
}
