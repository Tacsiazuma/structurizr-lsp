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
- [x] Inlay hint on name, description and technology
- [x] Document formatting
- [ ] Semantic analysis based on the specs
- [ ] Handle cancel request
- [ ] Textdocument/hover 
- [ ] Go to definition
- [ ] Go to references
- [ ] Rename support
- [ ] Debounce diagnostic notifications

### Supported language elements

- [x] workspace
- [x] properties
- [x] !identifiers
- [x] !docs
- [x] !adrs
- [x] configuration
- [x] scope
- [x] visibility
- [x] users
