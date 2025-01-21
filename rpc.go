package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Notification struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
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

type Rpc struct {
	input  *bufio.Reader
	output *bufio.Writer
}

func NewRpc(input io.Reader, output io.Writer) *Rpc {
	return &Rpc{input: bufio.NewReader(input), output: bufio.NewWriter(output)}
}

// Read a single LSP message from stdin.
func (r *Rpc) readMessage() (string, error) {

	// Read headers
	var contentLength int
	for {
		line, err := r.input.ReadString('\n')
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
	_, err := r.input.Read(body)
	if err != nil {
		return "", err
	}

	// logger.Printf("Incoming request: %s\n", body)
	return string(body), nil
}

// Write an LSP response to stdout.
func (r *Rpc) writeMessage(response interface{}) error {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return err
	}

	content := fmt.Sprintf(
		"Content-Length: %d\n\n%s",
		len(jsonResponse),
		jsonResponse,
	)
	// logger.Printf("Outgoing response: %s\n", jsonResponse)
	_, err = r.output.Write([]byte(content))
	r.output.Flush()
	return err
}
