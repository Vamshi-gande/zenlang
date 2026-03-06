package lexer

import "fmt"

// Position holds a line and column number in the source file.
// Used to point the developer to exactly where a lex error occurred.
type Position struct {
	Line   int
	Column int
}

// LexerError is a position-aware error produced when the lexer encounters
// a character it cannot tokenize. It implements the standard error interface
// so it can be returned anywhere a Go error is expected.
type LexerError struct {
	Position Position
	Message  string
	Char     byte
}

// Error implements the error interface.
func (e *LexerError) Error() string {
	return fmt.Sprintf(
		"lexer error at line %d, column %d: %s",
		e.Position.Line,
		e.Position.Column,
		e.Message,
	)
}

// NewLexerError builds a LexerError for an unexpected character.
// Line and column tracking will be wired into InputReader during Phase 8.
func NewLexerError(line, column int, char byte) *LexerError {
	return &LexerError{
		Position: Position{Line: line, Column: column},
		Message:  fmt.Sprintf("unexpected character '%c'", char),
		Char:     char,
	}
}
