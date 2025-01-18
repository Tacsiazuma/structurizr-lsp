package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Request struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Notification struct {
	Method  string          `json:"method"`
	Params  interface{} `json:"params"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Diagnostic struct {
	Range   Range    `json:"range"`
	Message string `json:"message"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type PublishDiagnosticsParams struct {
	URI         string        `json:"uri"`
	Diagnostics []*Diagnostic `json:"diagnostics"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
    ContentChanges []ContentChange
}

type ContentChange struct {
    Text string `json:"text"`
}

var logger *log.Logger

func initLogger() {
	logFile, err := os.OpenFile("/home/tacsiazuma/work/structurizr-lsp/lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}

// Read a single LSP message from stdin.
func readMessage() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Read headers
	var contentLength int
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)

		if line == "" { // Empty line signals the end of headers
			break
		}

		if strings.HasPrefix(line, "Content-Length:") {
			fmt.Sscanf(line, "Content-Length: %d", &contentLength)
		}
	}

	// Read the body
	body := make([]byte, contentLength)
	_, err := reader.Read(body)
	if err != nil {
		return "", err
	}

	logger.Printf("Incoming request: %s\n", body)
	return string(body), nil
}

// Write an LSP response to stdout.
func writeMessage(response Response) error {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return err
	}

	content := fmt.Sprintf(
		"Content-Length: %d\r\n\r\n%s",
		len(jsonResponse),
		jsonResponse,
	)
	logger.Printf("Outgoing response: %s\n", jsonResponse)
	_, err = os.Stdout.Write([]byte(content))
	return err
}

// Write an LSP Notification to stdout.
func writeNotification(notif Notification) error {
	jsonResponse, err := json.Marshal(notif)
	if err != nil {
		return err
	}

	content := fmt.Sprintf(
		"Content-Length: %d\r\n\r\n%s",
		len(jsonResponse),
		jsonResponse,
	)
	logger.Printf("Outgoing notification: %s\n", jsonResponse)
	_, err = os.Stdout.Write([]byte(content))
	return err
}
