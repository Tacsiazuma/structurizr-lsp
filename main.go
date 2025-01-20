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
	rpc *Rpc
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

func (l *Lsp) Handle() {
	// Read message from client
	msg, err := l.rpc.readMessage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read message: %v\n", err)
		return
	}

	// Parse the JSON-RPC request
	var req Request
	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse JSON: %v\n", err)
		return
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
			fmt.Fprintf(os.Stderr, "Failed to parse 'didChange' params: %v\n", err)
			break
		}
		l.handleDidChange(params)
	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse 'didOpen' params: %v\n", err)
			break
		}
		l.handleDidOpen(params)
	case "shutdown":
		l.handleShutdown(req)
	default:
		l.sendError(req.ID, -32601, "Method not found")
	}
}

func initLogger() {
	logFile, err := os.OpenFile("lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}
