package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

var logger *log.Logger

func main() {
	initLogger()
	if os.Args[len(os.Args)-1] == "version" {
		info, _ := debug.ReadBuildInfo()
		fmt.Println(info.Main.Version)
		return
	}
	// Defer a function to recover from panics
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Recovered from panic: %v\n", r)
		}
	}()
	for {
		// Read message from client
		msg, err := readMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read message: %v\n", err)
			continue
		}

		// Parse the JSON-RPC request
		var req Request
		if err := json.Unmarshal([]byte(msg), &req); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse JSON: %v\n", err)
			continue
		}

		// Handle the request
		switch req.Method {
		case "initialize":
			handleInitialize(req)
		case "initialized": // notification does not require response
			break
		case "textDocument/didSave": // notification does not require response
			break
		case "textDocument/didChange":
			var params DidChangeTextDocumentParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				fmt.Printf("Failed to parse 'didChange' params: %v\n", err)
				break
			}
			handleDidChange(params)
		case "textDocument/didOpen":
			var params DidOpenTextDocumentParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				fmt.Printf("Failed to parse 'didOpen' params: %v\n", err)
				break
			}
			handleDidOpen(params)
		case "shutdown":
			handleShutdown(req)
		default:
			sendError(req.ID, -32601, "Method not found")
		}
	}
}

func initLogger() {
	logFile, err := os.OpenFile("lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}
