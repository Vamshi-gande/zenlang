package lexer

import (
	"github.com/Vamshi-gande/zenlang/pkg/token"
)

// Lexer tokenizes Zen source code
type Lexer struct {
	reader      *InputReader
	currentChar byte
}

func NewLexer(source string) *Lexer {
	l := &Lexer{}
	l.reader = NewInputReader(source)
	l.currentChar = l.reader.CurrentChar()
	return l
}

// NextToken is the only method the parser calls.
// It skips whitespace, identifies the current token, and returns it.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.currentChar {
	// Single character tokens
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: string(l.currentChar)}
		l.advance()
	case '-':
		tok = token.Token{Type: token.MINUS, Literal: string(l.currentChar)}
		l.advance()
	case '*':
		tok = token.Token{Type: token.ASTERISK, Literal: string(l.currentChar)}
		l.advance()
	case '/':
		tok = token.Token{Type: token.SLASH, Literal: string(l.currentChar)}
		l.advance()
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: string(l.currentChar)}
		l.advance()
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: string(l.currentChar)}
		l.advance()
	case '{':
		tok = token.Token{Type: token.LBRACE, Literal: string(l.currentChar)}
		l.advance()
	case '}':
		tok = token.Token{Type: token.RBRACE, Literal: string(l.currentChar)}
		l.advance()
	case '[':
		tok = token.Token{Type: token.LBRACKET, Literal: string(l.currentChar)}
		l.advance()
	case ']':
		tok = token.Token{Type: token.RBRACKET, Literal: string(l.currentChar)}
		l.advance()
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: string(l.currentChar)}
		l.advance()
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Literal: string(l.currentChar)}
		l.advance()

	// Two character tokens — require peeking
	case '=':
		if l.reader.PeekChar() == '=' {
			l.advance() // consume second '='
			tok = token.Token{Type: token.EQ, Literal: "=="}
		} else {
			tok = token.Token{Type: token.ASSIGN, Literal: "="}
		}
		l.advance()
	case '!':
		if l.reader.PeekChar() == '=' {
			l.advance() // consume '='
			tok = token.Token{Type: token.NOT_EQ, Literal: "!="}
		} else {
			tok = token.Token{Type: token.BANG, Literal: "!"}
		}
		l.advance()
	case '<':
		if l.reader.PeekChar() == '=' {
			l.advance()
			tok = token.Token{Type: token.LTE, Literal: "<="}
		} else {
			tok = token.Token{Type: token.LT, Literal: "<"}
		}
		l.advance()
	case '>':
		if l.reader.PeekChar() == '=' {
			l.advance()
			tok = token.Token{Type: token.GTE, Literal: ">="}
		} else {
			tok = token.Token{Type: token.GT, Literal: ">"}
		}
		l.advance()

	// EOF — InputReader returns null byte when past the end
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}

	// Multi-character tokens: identifiers, keywords, integers
	default:
		if isLetter(l.currentChar) {
			// First char is a letter/underscore — continue consuming letters, digits, underscores
			literal := l.readWhile(isIdentChar)
			tokenType := token.LookupIdentifier(literal)
			// readWhile already advanced past the last char — return directly
			return token.Token{Type: tokenType, Literal: literal}
		} else if isDigit(l.currentChar) {
			literal := l.readWhile(isDigit)
			return token.Token{Type: token.INT, Literal: literal}
		} else {
			// Unrecognized character
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.currentChar)}
			l.advance()
		}
	}

	return tok
}

// advance moves the reader forward by one character and updates currentChar.
func (l *Lexer) advance() {
	l.currentChar = l.reader.Advance()
}

// skipWhitespace consumes all leading whitespace before the next token.
func (l *Lexer) skipWhitespace() {
	for l.currentChar == ' ' || l.currentChar == '\t' ||
		l.currentChar == '\n' || l.currentChar == '\r' {
		l.advance()
	}
}

// readWhile consumes characters as long as condition holds, returning the accumulated string.
// After this call, currentChar is the first character that did NOT satisfy condition.
func (l *Lexer) readWhile(condition func(byte) bool) string {
	start := l.reader.position
	for condition(l.currentChar) {
		l.advance()
	}
	return l.reader.source[start:l.reader.position]
}

// isLetter returns true for a-z, A-Z, and underscore.
// Used to validate the FIRST character of an identifier.
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// isIdentChar returns true for letters, underscores, AND digits.
// Used for every character AFTER the first in an identifier — allows counter_1, myVar2 etc.
func isIdentChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch)
}

// isDigit returns true for 0-9.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
