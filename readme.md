# Structurizr LSP

A language server for Structurizr written in Go.

### Usage

Build it with 

```
go build
```

Then set it up as the executable to run by your IDE using stdin/stdout as an interface:

```
./structurizr-lsp
```

### Known issues
- If a directory is included the parsed tokens contain the folder as their source instead of the actual file

### TODO

- [x] When problems are solved in a file push empty slice of diagnostics
- [ ] Semantic analysis based on the specs
- [ ] Document formatting
- [ ] Inlay hint on name, description and technology
- [ ] Handle cancel request
- [ ] Textdocument/hover 
- [ ] Go to definition
- [ ] Go to references
- [ ] Rename support
- [ ] Debounce diagnostic notifications
