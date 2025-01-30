package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/tacsiazuma/structurizr-lsp/lsp"
)

func main() {
	logger := initLogger()
	if os.Args[len(os.Args)-1] == "version" {
		info, _ := debug.ReadBuildInfo()
		fmt.Println(info.Main.Version)
		return
	}
	lsp := lsp.From(os.Stdin, os.Stdout, logger)
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

func initLogger() *log.Logger {
	logFile, err := os.OpenFile("lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	return log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}
