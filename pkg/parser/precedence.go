package parser

import "github.com/Vamshi-gande/zenlang/pkg/token"

// Precedence levels. Higher number = tighter binding.
const (
	_           int = iota // 0 — unused
	LOWEST                 // 1 — weakest, starting point for any expression
	EQUALS                 // 2 — == !=
	LESSGREATER            // 3 — < > <= >=
	SUM                    // 4 — + -
	PRODUCT                // 5 — * /
	PREFIX                 // 6 — -x  !x  ++x  --x
	CALL                   // 7 — fn(args)
	INDEX                  // 8 — arr[0]
)

// precedences maps each infix/postfix token type to its precedence level.
// The parser looks up the peek token here to decide whether to keep consuming.
var precedences = map[token.TokenType]int{
	token.ASSIGN:          LOWEST + 1, // bare assignment: x = expr  (weaker than everything else)
	token.EQ:              EQUALS,
	token.NOT_EQ:          EQUALS,
	token.LT:              LESSGREATER,
	token.GT:              LESSGREATER,
	token.LTE:             LESSGREATER,
	token.GTE:             LESSGREATER,
	token.PLUS:            SUM,
	token.MINUS:           SUM,
	token.SLASH:           PRODUCT,
	token.ASTERISK:        PRODUCT,
	token.AND:             EQUALS, // && binds like equality
	token.OR:              EQUALS, // || binds like equality
	token.PLUS_ASSIGN:     EQUALS, // compound assignment, lowest among operators
	token.MINUS_ASSIGN:    EQUALS,
	token.ASTERISK_ASSIGN: EQUALS,
	token.SLASH_ASSIGN:    EQUALS,
	token.LPAREN:          CALL,
	token.LBRACKET:        INDEX,
}

// peekPrecedence returns the precedence of the parser's peek token.
// Returns LOWEST if the token has no entry in the table.
func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// currentPrecedence returns the precedence of the parser's current token.
// Returns LOWEST if the token has no entry in the table.
func (p *Parser) currentPrecedence() int {
	if prec, ok := precedences[p.currentToken.Type]; ok {
		return prec
	}
	return LOWEST
}
