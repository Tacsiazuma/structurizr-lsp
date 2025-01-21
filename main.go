package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
)

var logger *log.Logger

type Lsp struct {
	initialized bool
	rpc         *Rpc
}

func From(input io.Reader, output io.Writer) *Lsp {
	r := NewRpc(input, output)
	return &Lsp{rpc: r}
}

func main() {
	initLogger()
	if os.Args[len(os.Args)-1] == "version" {
		info, _ := debug.ReadBuildInfo()
		fmt.Println(info.Main.Version)
		return
	}
	lsp := From(os.Stdin, os.Stdout)
	// Defer a function to recover from panics
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Recovered from panic: %v\n", r)
		}
	}()
	for {
		lsp.Handle()
	}
}

func (l *Lsp) Handle() error {
	// Read message from client
	msg, err := l.rpc.readMessage()
	if err != nil {
		return fmt.Errorf("Failed to read message: %v", err)
	}

	// Parse the JSON-RPC request
	var req Request
	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		return fmt.Errorf("Failed to parse JSON: %v", err)
	}

	// Handle the request
	switch req.Method {
	case "initialize":
		l.handleInitialize(req)
	case "initialized": // notification does not require response
		break
	case "textDocument/didSave": // notification does not require response
		break
	case "textDocument/completion": // not implemented yet
		break
	case "$/cancellation": // not implemented yet
		break
	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return fmt.Errorf("Failed to parse 'didChange' params: %v", err)
		}
		l.handleDidChange(params)
	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return fmt.Errorf("Failed to parse 'didOpen' params: %v", err)
		}
		l.handleDidOpen(params)
	case "shutdown":
		if l.initialized {
			l.handleShutdown(req)
		} else {
			l.sendError(req.ID, -32002, "Not initialized")
		}
	default:
		l.sendError(req.ID, -32601, "Method not found")
	}
	return nil
}

func initLogger() {
	logFile, err := os.OpenFile("lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}
