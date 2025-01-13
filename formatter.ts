export default class Formatter {
  /**
   * Formats Structurizr DSL content with proper indentation and spacing.
   * @param dsl - The Structurizr DSL content to format.
   * @returns An object containing the formatted DSL, last line number, and the length of the last line.
   */
  static formatStructurizrDSL(dsl: string): { formatted: string; lastLineNumber: number; lastLineLength: number } {
    if (!dsl || typeof dsl !== 'string') {
      throw new Error('DSL content must be a non-empty string');
    }

    let indentLevel = 0;
    const indentSize = 2;

    const lines = dsl
      .replace(/\r\n|\r/g, '\n') // Normalize line endings to \n
      .split('\n') // Split into lines
      .map(line => {
        const trimmed = line.trim();

        // Adjust indentation level based on brackets
        if (trimmed.endsWith('}')) {
          indentLevel = Math.max(0, indentLevel - 1);
        }

        const indentedLine = ' '.repeat(indentLevel * indentSize) + trimmed;

        if (trimmed.endsWith('{')) {
          indentLevel++;
        }

        return indentedLine;
      })
      .filter(line => line.length > 0); // Remove empty lines

    const formatted = lines.join('\n');
    const lastLine = lines[lines.length - 1] || '';

    return {
      formatted,
      lastLineNumber: lines.length,
      lastLineLength: lastLine.length
    };
  }
}
