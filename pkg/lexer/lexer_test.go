package lexer

import (
	"testing"

	"github.com/Vamshi-gande/zenlang/pkg/token"
)

// expectedToken holds what a single NextToken() call should produce.
type expectedToken struct {
	expectedType    token.TokenType
	expectedLiteral string
}

// runLexerTest drives a full token sequence comparison.
// On failure it tells you exactly which token index was wrong and what was expected vs received.
func runLexerTest(t *testing.T, input string, expected []expectedToken) {
	t.Helper()
	l := NewLexer(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.expectedType {
			t.Errorf("token[%d]: expected type=%q, got type=%q (literal=%q)",
				i, exp.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.expectedLiteral {
			t.Errorf("token[%d]: expected literal=%q, got literal=%q (type=%q)",
				i, exp.expectedLiteral, tok.Literal, tok.Type)
		}
	}
}

// ---------------------------------------------------------------------------
// Single character tokens
// ---------------------------------------------------------------------------

func TestSingleCharTokens(t *testing.T) {
	input := "+ - * / ( ) { } [ ] , ;"
	expected := []expectedToken{
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	runLexerTest(t, input, expected)
}

// ---------------------------------------------------------------------------
// Two character tokens — exercises the peek logic
// ---------------------------------------------------------------------------

func TestTwoCharTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []expectedToken
	}{
		{"==", []expectedToken{{token.EQ, "=="}, {token.EOF, ""}}},
		{"!=", []expectedToken{{token.NOT_EQ, "!="}, {token.EOF, ""}}},
		{"<=", []expectedToken{{token.LTE, "<="}, {token.EOF, ""}}},
		{">=", []expectedToken{{token.GTE, ">="}, {token.EOF, ""}}},
		// Single char variants — peek must NOT over-consume
		{"=", []expectedToken{{token.ASSIGN, "="}, {token.EOF, ""}}},
		{"!", []expectedToken{{token.BANG, "!"}, {token.EOF, ""}}},
		{"<", []expectedToken{{token.LT, "<"}, {token.EOF, ""}}},
		{">", []expectedToken{{token.GT, ">"}, {token.EOF, ""}}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runLexerTest(t, tt.input, tt.expected)
		})
	}
}

// ---------------------------------------------------------------------------
// Whitespace is ignored
// ---------------------------------------------------------------------------

func TestWhitespaceHandling(t *testing.T) {
	compact := "let x = 5"
	spaced := "let   x   =   5"
	tabbed := "let\tx\t=\t5"
	newlined := "let\nx\n=\n5"

	expected := []expectedToken{
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.EOF, ""},
	}

	for _, input := range []string{compact, spaced, tabbed, newlined} {
		t.Run(input, func(t *testing.T) {
			runLexerTest(t, input, expected)
		})
	}
}

// ---------------------------------------------------------------------------
// Integer literals — multi-digit must be one token
// ---------------------------------------------------------------------------

func TestIntegerLiterals(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"5", "5"},
		{"0", "0"},
		{"123", "123"},
		{"9999", "9999"},
	}

	for _, tt := range tests {
		expected := []expectedToken{
			{token.INT, tt.literal},
			{token.EOF, ""},
		}
		runLexerTest(t, tt.input, expected)
	}
}

// ---------------------------------------------------------------------------
// Identifiers — including underscores
// ---------------------------------------------------------------------------

func TestIdentifiers(t *testing.T) {
	tests := []struct{ input, literal string }{
		{"x", "x"},
		{"myVar", "myVar"},
		{"counter_1", "counter_1"},
		{"_private", "_private"},
	}

	for _, tt := range tests {
		expected := []expectedToken{
			{token.IDENT, tt.literal},
			{token.EOF, ""},
		}
		runLexerTest(t, tt.input, expected)
	}
}

// ---------------------------------------------------------------------------
// Keywords must not come back as IDENT
// ---------------------------------------------------------------------------

func TestKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected token.TokenType
	}{
		{"let", token.LET},
		{"fn", token.FUNCTION},
		{"if", token.IF},
		{"else", token.ELSE},
		{"return", token.RETURN},
		{"while", token.WHILE},
		{"true", token.TRUE},
		{"false", token.FALSE},
	}

	for _, tt := range tests {
		expected := []expectedToken{
			{tt.expected, tt.input},
			{token.EOF, ""},
		}
		runLexerTest(t, tt.input, expected)
	}
}

// ---------------------------------------------------------------------------
// Keywords embedded inside longer identifiers must NOT match
// ---------------------------------------------------------------------------

func TestKeywordsInsideIdentifiers(t *testing.T) {
	tests := []string{"letter", "ifelse", "returned", "truthy", "fnord", "letting"}

	for _, input := range tests {
		expected := []expectedToken{
			{token.IDENT, input},
			{token.EOF, ""},
		}
		runLexerTest(t, input, expected)
	}
}

// ---------------------------------------------------------------------------
// Full statement — let binding
// ---------------------------------------------------------------------------

func TestLetStatement(t *testing.T) {
	input := `let x = 10;`
	expected := []expectedToken{
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	runLexerTest(t, input, expected)
}

// ---------------------------------------------------------------------------
// Full statement — function definition
// ---------------------------------------------------------------------------

func TestFunctionDefinition(t *testing.T) {
	input := `let add = fn(a, b) { return a + b; }`
	expected := []expectedToken{
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.IDENT, "a"},
		{token.PLUS, "+"},
		{token.IDENT, "b"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}
	runLexerTest(t, input, expected)
}

// ---------------------------------------------------------------------------
// Conditional statement
// ---------------------------------------------------------------------------

func TestConditional(t *testing.T) {
	input := `if (x == y) { return true; } else { return false; }`
	expected := []expectedToken{
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.EQ, "=="},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}
	runLexerTest(t, input, expected)
}

// ---------------------------------------------------------------------------
// EOF handling — must not panic on repeated calls
// ---------------------------------------------------------------------------

func TestEOFHandling(t *testing.T) {
	// Empty string → immediate EOF
	runLexerTest(t, "", []expectedToken{{token.EOF, ""}})

	// After last real token, EOF must be returned repeatedly without panic
	l := NewLexer("5")
	l.NextToken()           // INT "5"
	first := l.NextToken()  // EOF
	second := l.NextToken() // EOF again — must not panic
	if first.Type != token.EOF {
		t.Errorf("expected EOF after last token, got %q", first.Type)
	}
	if second.Type != token.EOF {
		t.Errorf("expected EOF on repeated call, got %q", second.Type)
	}
}

// ---------------------------------------------------------------------------
// Illegal characters — lexer must recover and continue
// ---------------------------------------------------------------------------

func TestIllegalCharacters(t *testing.T) {
	// Isolated illegal characters
	runLexerTest(t, "@", []expectedToken{{token.ILLEGAL, "@"}, {token.EOF, ""}})
	runLexerTest(t, "$", []expectedToken{{token.ILLEGAL, "$"}, {token.EOF, ""}})

	// Illegal in the middle — lexer must keep going after it
	input := "x @ y"
	expected := []expectedToken{
		{token.IDENT, "x"},
		{token.ILLEGAL, "@"},
		{token.IDENT, "y"},
		{token.EOF, ""},
	}
	runLexerTest(t, input, expected)
}
