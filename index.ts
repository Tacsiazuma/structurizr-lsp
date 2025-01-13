import fs from 'fs';
import Formatter from './formatter';
import { DeclarationParams, DeclarationResult, FormattingParams } from './contract';
import { log } from 'console';
// Create a writable log file
const logStream = fs.createWriteStream('language_server.log', { flags: 'a' });
console.log("Language Server is running. Logging all stdin data to language_server.log...");
process.stdin.setEncoding('utf8');
process.stdin.on('data', (chunk: string) => {
    const jason = JSON.parse(chunk.split("\r\n\r\n")[1])
    logStream.write(`Received: ${chunk}\r\n\r\n`);
    if (jason.method == 'initialized') {
        return
    }
    if (jason.method == 'shutdown') {
        return process.exit()
    }
    if (jason.method == 'textDocument/didSave') {
        return
    }
    if (jason.method == 'textDocument/formatting') {
        return formatting(jason.id, jason.params)
    }
    if (jason.method == 'textDocument/definition') {
        return declaration(jason.id, jason.params)
    }
    const response = JSON.stringify({
        id: jason.id,
        result: {
            capabilities: {
                documentFormattingProvider: true,
                definitionProvider: true
            },
            serverInfo: {
                version: "1.0.0",
                name: "structurizr LSP"
            }
        }
    })
    process.stdout.write("Content-Length: " + response.length + "\r\n\r\n")
    logStream.write("Sent: " + response + "\r\n\r\n")
    process.stdout.write(response)
});
process.stdin.on('end', () => {
    logStream.write("Language Server shutting down.\n");
    logStream.end();
    console.log("Language Server shutting down.");
});
function declaration(id: number, params: DeclarationParams) {
    const content = fs.readFileSync(params.textDocument.uri.slice(7))
    const lines = content.toString().split("\n")
    const word = getWordAtPosition(lines[params.position.line], params.position.character)
    const result = findDeclaration(lines, word)
    if (!result) {
        return
    }
    logStream.write("Getting declaration for " + word + "\r\n\r\n")
    const response = JSON.stringify({
        id: id,
        result: { uri: params.textDocument.uri, range: { start: { line: result.line, character: result.position } } }
    } as DeclarationResult)
    process.stdout.write("Content-Length: " + response.length + "\r\n\r\n")
    logStream.write("Sent: " + response + "\r\n\r\n")
    process.stdout.write(response)
}
function findDeclaration(lines: string[], searchString: string): { line: number; position: number } | null {
    const regex = new RegExp(`^\\s*${searchString}\\s*=`); // Match only if the string is preceded by spaces or starts the line

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        const match = regex.exec(line);
        if (match) {
            return { line: i, position: line.indexOf(searchString) };
        }
    }

    return null; // Return null if no match is found
}
/**
   * Extracts a word from a line at a specific character position surrounded by spaces.
   * @param line - The line of text to search within.
   * @param charPosition - The character position to start the search.
   * @returns The word found at the specified position, or an empty string if none found.
   */
function getWordAtPosition(line: string, charPosition: number): string {
    if (typeof line !== 'string' || typeof charPosition !== 'number') {
        throw new Error('Invalid arguments. Expected a string and a number.');
    }

    if (charPosition < 0 || charPosition >= line.length) {
        return '';
    }

    const left = line.slice(0, charPosition).lastIndexOf(' ');
    const right = line.slice(charPosition).indexOf(' ');

    const start = left === -1 ? 0 : left + 1;
    const end = right === -1 ? line.length : charPosition + right;

    return line.slice(start, end).trim();
}

function formatting(id: number, params: FormattingParams) {
    const content = fs.readFileSync(params.textDocument.uri.slice(7))
    const { formatted, lastLineNumber, lastLineLength } = Formatter.formatStructurizrDSL(content.toString())
    const response = JSON.stringify({
        id: id,
        result: [
            {
                newText: formatted,
                range: {
                    start: { line: 0, character: 0 },
                    end: { line: lastLineNumber, character: lastLineLength }
                }
            }
        ]
    })
    process.stdout.write("Content-Length: " + response.length + "\r\n\r\n")
    logStream.write("Sent: " + response + "\r\n\r\n")
    process.stdout.write(response)
}
