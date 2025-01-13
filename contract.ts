
/**
 * Value-object describing what options formatting should use.
 */
export interface FormattingOptions {
    /**
     * Size of a tab in spaces.
     */
    tabSize: number;

    /**
     * Prefer spaces over tabs.
     */
    insertSpaces: boolean;

    /**
     * Trim trailing whitespace on a line.
     *
     * @since 3.15.0
     */
    trimTrailingWhitespace?: boolean;

    /**
     * Insert a newline character at the end of the file if one does not exist.
     *
     * @since 3.15.0
     */
    insertFinalNewline?: boolean;

    /**
     * Trim all newlines after the final newline at the end of the file.
     *
     * @since 3.15.0
     */
    trimFinalNewlines?: boolean;

}
export interface TextDocument {
    uri: string
}

export interface FormattingParams {
    textDocument: TextDocument
    options: FormattingParams
}

export interface Position {
    line: number
    character: number
}
export interface DeclarationParams {
    textDocument: TextDocument
    position: Position
}

export interface Range {
    start: Position
    end: Position
}

export interface Location {
    uri: string
    range: Range
}

export interface DeclarationResult {
    id : number,
    result : Location | Location[] | null
}
