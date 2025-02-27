package lsp

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tacsiazuma/structurizr-lsp/rpc"
)

func (l *Lsp) handleFormatting(id int, param FormattingParams) {
	content, err := l.getOrUpdateContent(param.TextDocument.URI, param.TextDocument.Text)
	if err != nil {
		l.sendError(id, 1, "Cannot format without content")
	}
	input := content.Text
	indentLevel := 0
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))
	var edits []TextEdit
	lineNum := 0

	for scanner.Scan() {
		originalLine := scanner.Text()
		line := strings.TrimSpace(originalLine)
		
		var formattedLine string
		if line == "" {
			formattedLine = "" // Preserve empty lines
		} else if strings.HasSuffix(line, "{") {
			formattedLine = strings.Repeat("    ", indentLevel) + line
			indentLevel++
		} else if strings.Contains(line, "}") {
			// Count closing braces to handle multiple on one line
			closingCount := strings.Count(line, "}")
			indentLevel -= closingCount
			if indentLevel < 0 {
				indentLevel = 0
			}
			formattedLine = strings.Repeat("    ", indentLevel) + line
		} else {
			formattedLine = strings.Repeat("    ", indentLevel) + line
		}

		sb.WriteString(formattedLine + "\n")
		// Always create an edit, even for empty lines
		edits = append(edits, TextEdit{
			Range: Range{
				Start: Position{Line: lineNum, Character: 0},
				End:   Position{Line: lineNum, Character: len(originalLine)},
			},
			NewText: formattedLine,
		})
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		l.logger.Println("error reading input: " + err.Error())
		return
	}
	l.registerContent(param.TextDocument.URI, sb.String(), content.Ast) // update the content after formatting
	response := rpc.Response{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  edits,
	}
	if err := l.rpc.WriteMessage(response); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send error response: %v\n", err)
	}
}
