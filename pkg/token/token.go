package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}
// Defining taoken types
const (
	ILLEGAL="ILLEGAL" // unknown token
	EOF="EOF"         // end of file

	// Identifiers + literals
	IDENT="IDENT" // add, foobar, x, y, ...
	LET="LET"
	INT="INT"
	FLOAT="FLOAT"
	STRING="STRING"
	TRUE="TRUE"
	FALSE="FALSE"
	NULL="NULL"

	// Operators
	ASSIGN="="
	PLUS="+"
	MINUS="-"
	BANG="!"
	ASTERISK="*"
	SLASH="/"
	LT="<"
	GT=">"
	EQ="=="
	NOT_EQ="!="
	LTE="<="
	GTE=">="
	AND="&&"
	OR="||"
	INC="++"
	DEC="--"
	PLUS_ASSIGN="+="
	MINUS_ASSIGN="-="
	ASTERISK_ASSIGN="*="
	SLASH_ASSIGN="/="


	// Delimiters
	COMMA=","
	SEMICOLON=";"
	COLON=":"
	LPAREN="("
	RPAREN=")"
	LBRACE="{"
	RBRACE="}"
	LBRACKET="["
	RBRACKET="]"
)