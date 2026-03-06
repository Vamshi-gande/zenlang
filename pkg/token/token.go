package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

// Defining token types
const (
	ILLEGAL = "ILLEGAL" // unknown token
	EOF     = "EOF"     // end of file

	// Identifiers + literals
	IDENT  = "IDENT"
	INT    = "INT"
	FLOAT  = "FLOAT"
	STRING = "STRING"

	// Keywords
	LET      = "LET"
	FUNCTION = "FUNCTION"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	NULL     = "NULL"
	RETURN   = "RETURN"
	IF       = "IF"
	ELSE     = "ELSE"
	WHILE    = "WHILE"

	// Operators
	ASSIGN          = "="
	PLUS            = "+"
	MINUS           = "-"
	BANG            = "!"
	ASTERISK        = "*"
	SLASH           = "/"
	LT              = "<"
	GT              = ">"
	EQ              = "=="
	NOT_EQ          = "!="
	LTE             = "<="
	GTE             = ">="
	AND             = "&&"
	OR              = "||"
	INC             = "++"
	DEC             = "--"
	PLUS_ASSIGN     = "+="
	MINUS_ASSIGN    = "-="
	ASTERISK_ASSIGN = "*="
	SLASH_ASSIGN    = "/="

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
)

// keywords maps reserved words to their token types.
// LookupIdentifier uses this to distinguish keywords from plain identifiers.
var keywords = map[string]TokenType{
	"let":    LET,
	"fn":     FUNCTION,
	"true":   TRUE,
	"false":  FALSE,
	"null":   NULL,
	"return": RETURN,
	"if":     IF,
	"else":   ELSE,
	"while":  WHILE,
}

// LookupIdentifier returns the keyword TokenType if ident is a reserved word,
// otherwise returns IDENT.
func LookupIdentifier(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
