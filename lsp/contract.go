package lsp

type Diagnostic struct {
	Range   Range  `json:"range"`
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

type InlayHintKind int

var (
	TypeKind      InlayHintKind = 1
	ParameterKind InlayHintKind = 2
)

type InlayHint struct {
	Position Position      `json:"position"`
	Label    string        `json:"label"`
	Kind     InlayHintKind `json:"kind"`
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

type InlayHintParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
	Range        Range            `json:"range"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   TextDocumentItem `json:"textDocument"`
	ContentChanges []ContentChange
}

type ContentChange struct {
	Text string `json:"text"`
}

type FormattingParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// FormatResponse represents the LSP format response structure.
type FormatResponse struct {
	Id     int        `json:"id"`
	Edits  []TextEdit `json:"edits"`
	Method string     `json:"method"`
}

// TextEdit represents a single text edit.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}
