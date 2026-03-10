package parser

import (
	"fmt"

	"github.com/Vamshi-gande/zenlang/pkg/token"
)

// ParseError holds the token where the problem occurred and a message describing
// what was expected versus what was found.
type ParseError struct {
	Token   token.Token
	Message string
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at '%s': %s", e.Token.Literal, e.Message)
}

// NewParseError constructs a ParseError with the given token and message.
func NewParseError(tok token.Token, msg string) *ParseError {
	return &ParseError{Token: tok, Message: msg}
}

// Errors returns all accumulated parse errors.
func (p *Parser) Errors() []*ParseError {
	return p.errors
}

// peekError records an error for when peekToken is not the expected type.
func (p *Parser) peekError(expected token.TokenType) {
	msg := fmt.Sprintf("expected next token to be '%s', got '%s' instead",
		expected, p.peekToken.Type)
	p.errors = append(p.errors, NewParseError(p.peekToken, msg))
}

// noPrefixParseFnError records an error for when no prefix parse function
// is registered for the current token — meaning the token cannot start an expression.
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function found for token type '%s'", t)
	p.errors = append(p.errors, NewParseError(p.currentToken, msg))
}
