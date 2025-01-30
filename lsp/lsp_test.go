package lsp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strings"
	"testing"
)

// TestCase holds the input and expected output for a test case.
type TestCase struct {
	Input  string
	Output string
}

func initLogger() *log.Logger {
	logFile, err := os.OpenFile("lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	return log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}
func TestRpc(t *testing.T) {
	logger := initLogger()
	t.Run("initialize", func(t *testing.T) {
		writer := &UnbufferedWriter{}
		reader := &StringReader{}

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
	})
	t.Run("textdocument/didOpen", func(t *testing.T) {
		writer := &UnbufferedWriter{}
		reader := &StringReader{}

		sut := From(reader, writer, logger)
		t.Run("results in publish diagnostics", func(t *testing.T) {
			testcase := ParseTestFile("textdocument_didopen", "publish_diagnostics")
			reader.SetString(testcase.Input)
			err := sut.Handle()
			assert.Nil(t, err)
			assert.Equal(t, testcase.Output, writer.written)
		})
	})
	t.Run("textdocument/inlayHint", func(t *testing.T) {
		writer := &UnbufferedWriter{}
		reader := &StringReader{}

		sut := From(reader, writer, logger)
		t.Run("returns inlay hints if file loaded", func(t *testing.T) {
			LoadFile(reader, writer, sut)
			testcase := ParseTestFile("textdocument_inlayhint", "textdocument_inlayhint")
			reader.SetString(testcase.Input)
			err := sut.Handle()
			assert.Nil(t, err)
			assert.Equal(t, testcase.Output, writer.written)
		})
	})
}

func LoadFile(reader *StringReader, writer *UnbufferedWriter, sut *Lsp) {
	c := ParseTestFile("textdocument_didopen", "publish_diagnostics")
	reader.SetString(c.Input)
	err := sut.Handle()
	if err != nil {
		log.Fatal(err)
	}
	writer.Reset()
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

func (w *UnbufferedWriter) Reset() {
	w.written = ""
}

func MinifyJSON(input string) (string, error) {
	var buffer bytes.Buffer
	err := json.Compact(&buffer, []byte(input))
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
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
	trimmedout, err := MinifyJSON(string(o))
	if err != nil {
		log.Fatal(err)
	}
	return &TestCase{
		Input:  fmt.Sprintf("Content-Length: %d\n\n%s", len(i), string(i)),
		Output: fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(trimmedout), trimmedout),
	}
}
