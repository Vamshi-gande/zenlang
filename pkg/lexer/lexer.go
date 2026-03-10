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
	// ---------------------------------------------------------------
	// Single character tokens
	// ---------------------------------------------------------------
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
	case ':':
		tok = token.Token{Type: token.COLON, Literal: string(l.currentChar)}
		l.advance()

	// ---------------------------------------------------------------
	// Two-character tokens — require peeking at the next character
	// ---------------------------------------------------------------
	case '=':
		if l.reader.PeekChar() == '=' {
			l.advance()
			tok = token.Token{Type: token.EQ, Literal: "=="}
		} else {
			tok = token.Token{Type: token.ASSIGN, Literal: "="}
		}
		l.advance()

	case '!':
		if l.reader.PeekChar() == '=' {
			l.advance()
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

	case '+':
		if l.reader.PeekChar() == '+' {
			l.advance()
			tok = token.Token{Type: token.INC, Literal: "++"}
		} else if l.reader.PeekChar() == '=' {
			l.advance()
			tok = token.Token{Type: token.PLUS_ASSIGN, Literal: "+="}
		} else {
			tok = token.Token{Type: token.PLUS, Literal: "+"}
		}
		l.advance()

	case '-':
		if l.reader.PeekChar() == '-' {
			l.advance()
			tok = token.Token{Type: token.DEC, Literal: "--"}
		} else if l.reader.PeekChar() == '=' {
			l.advance()
			tok = token.Token{Type: token.MINUS_ASSIGN, Literal: "-="}
		} else {
			tok = token.Token{Type: token.MINUS, Literal: "-"}
		}
		l.advance()

	case '*':
		if l.reader.PeekChar() == '=' {
			l.advance()
			tok = token.Token{Type: token.ASTERISK_ASSIGN, Literal: "*="}
		} else {
			tok = token.Token{Type: token.ASTERISK, Literal: "*"}
		}
		l.advance()

	case '/':
		if l.reader.PeekChar() == '=' {
			l.advance()
			tok = token.Token{Type: token.SLASH_ASSIGN, Literal: "/="}
		} else {
			tok = token.Token{Type: token.SLASH, Literal: "/"}
		}
		l.advance()

	case '&':
		if l.reader.PeekChar() == '&' {
			l.advance()
			tok = token.Token{Type: token.AND, Literal: "&&"}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.currentChar)}
		}
		l.advance()

	case '|':
		if l.reader.PeekChar() == '|' {
			l.advance()
			tok = token.Token{Type: token.OR, Literal: "||"}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.currentChar)}
		}
		l.advance()

	// ---------------------------------------------------------------
	// String literals
	// ---------------------------------------------------------------
	case '"':
		literal := l.readString()
		return token.Token{Type: token.STRING, Literal: literal}

	// ---------------------------------------------------------------
	// EOF
	// ---------------------------------------------------------------
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}

	// ---------------------------------------------------------------
	// Multi-character tokens: identifiers, keywords, numbers
	// ---------------------------------------------------------------
	default:
		if isLetter(l.currentChar) {
			literal := l.readWhile(isIdentChar)
			tokenType := token.LookupIdentifier(literal)
			return token.Token{Type: tokenType, Literal: literal}
		} else if isDigit(l.currentChar) {
			return l.readNumber()
		} else {
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

// readNumber reads an integer or float literal.
// If a '.' is encountered after digits, it continues consuming as a float.
func (l *Lexer) readNumber() token.Token {
	start := l.reader.position
	for isDigit(l.currentChar) {
		l.advance()
	}
	// check for decimal point followed by more digits → float
	if l.currentChar == '.' && isDigit(l.reader.PeekChar()) {
		l.advance() // consume '.'
		for isDigit(l.currentChar) {
			l.advance()
		}
		literal := l.reader.source[start:l.reader.position]
		return token.Token{Type: token.FLOAT, Literal: literal}
	}
	literal := l.reader.source[start:l.reader.position]
	return token.Token{Type: token.INT, Literal: literal}
}

// readString consumes characters between double quotes and returns the inner content.
// currentChar must be '"' on entry. The closing '"' is consumed before returning.
func (l *Lexer) readString() string {
	l.advance() // consume opening '"'
	start := l.reader.position
	for l.currentChar != '"' && l.currentChar != 0 {
		l.advance()
	}
	str := l.reader.source[start:l.reader.position]
	l.advance() // consume closing '"'
	return str
}

// isLetter returns true for a-z, A-Z, and underscore.
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// isIdentChar returns true for letters, underscores, AND digits.
func isIdentChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch)
}

// isDigit returns true for 0-9.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
