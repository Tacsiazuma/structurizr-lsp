package lsp

import (
	"bytes"
	"errors"
	"fmt"
	"log"
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
	writer := &UnbufferedWriter{}
	reader := &StringReader{}
	logger := log.New(&bytes.Buffer{}, "", log.LstdFlags|log.Lshortfile)

	sut := From(reader, writer, logger)
	t.Run("request return error if not initialized first", func(t *testing.T) {
		testcase := ParseTestFile("shutdown", "unsuccessful_initialize")
		reader.SetString(testcase.Input)
		err := sut.Handle()
		assert.Nil(t, err)
		assert.Equal(t, testcase.Output, writer.written)
	})
	t.Run("initialize successful", func(t *testing.T) {
		testcase := ParseTestFile("initialize", "successful_initialize")
		reader.SetString(testcase.Input)
		sut.Handle()
		assert.Equal(t, testcase.Output, writer.written)
	})
	t.Run("textdocument/didOpen", func(t *testing.T) {
		t.Run("results in publish diagnostics", func(t *testing.T) {
			testcase := ParseTestFile("textdocument_didopen", "publish_diagnostics")
			reader.SetString(testcase.Input)
			err := sut.Handle()
			assert.Nil(t, err)
			assert.Equal(t, testcase.Output, writer.written)
		})
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
func ParseTestFile(input, output string) *TestCase {
	// Read the file contents.
	i, err := os.ReadFile("../fixture/input/" + input + ".json")
	if err != nil {
		log.Fatal(err)
	}
	o, err := os.ReadFile("../fixture/output/" + output + ".json")
	if err != nil {
		log.Fatal(err)
	}
	// we need to trim output for later assertion
	trimmedout := strings.TrimRight(string(o), "\n")
	return &TestCase{
		Input:  fmt.Sprintf("Content-Length: %d\n\n%s", len(i), string(i)),
		Output: fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(trimmedout), trimmedout),
	}
}
