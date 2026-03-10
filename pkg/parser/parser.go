package parser

import (
	"fmt"
	"strconv"

	"github.com/Vamshi-gande/zenlang/pkg/ast"
	"github.com/Vamshi-gande/zenlang/pkg/lexer"
	"github.com/Vamshi-gande/zenlang/pkg/token"
)

// prefixParseFn is called when the associated token type appears at the start
// of an expression (prefix position). Returns the parsed expression.
type prefixParseFn func() ast.Expression

// infixParseFn is called when the associated token type appears between two
// expressions (infix position). Receives the already-parsed left expression.
type infixParseFn func(ast.Expression) ast.Expression

// Parser consumes a token stream from the lexer and produces an AST.
type Parser struct {
	lexer        *lexer.Lexer
	currentToken token.Token
	peekToken    token.Token
	errors       []*ParseError

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// NewParser creates a parser ready to parse the given lexer's token stream.
// It registers all parse functions and primes the two-token lookahead window.
func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{lexer: l}
	p.registerParseFunctions()
	p.nextToken() // primes peekToken
	p.nextToken() // primes currentToken
	return p
}

// registerParseFunctions maps every token type to its prefix and/or infix
// parse function. Called once during construction.
func (p *Parser) registerParseFunctions() {
	// --- prefix functions ---
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.prefixParseFns[token.IDENT] = p.parseIdentifier
	p.prefixParseFns[token.INT] = p.parseIntegerLiteral
	p.prefixParseFns[token.FLOAT] = p.parseFloatLiteral
	p.prefixParseFns[token.STRING] = p.parseStringLiteral
	p.prefixParseFns[token.TRUE] = p.parseBoolean
	p.prefixParseFns[token.FALSE] = p.parseBoolean
	p.prefixParseFns[token.NULL] = p.parseNullLiteral
	p.prefixParseFns[token.BANG] = p.parsePrefixExpression
	p.prefixParseFns[token.MINUS] = p.parsePrefixExpression
	p.prefixParseFns[token.INC] = p.parsePrefixExpression
	p.prefixParseFns[token.DEC] = p.parsePrefixExpression
	p.prefixParseFns[token.LPAREN] = p.parseGroupedExpression
	p.prefixParseFns[token.IF] = p.parseIfExpression
	p.prefixParseFns[token.FUNCTION] = p.parseFunctionLiteral
	p.prefixParseFns[token.LBRACKET] = p.parseArrayLiteral
	p.prefixParseFns[token.LBRACE] = p.parseHashLiteral

	// --- infix functions ---
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.infixParseFns[token.ASSIGN] = p.parseInfixExpression
	p.infixParseFns[token.PLUS] = p.parseInfixExpression
	p.infixParseFns[token.MINUS] = p.parseInfixExpression
	p.infixParseFns[token.SLASH] = p.parseInfixExpression
	p.infixParseFns[token.ASTERISK] = p.parseInfixExpression
	p.infixParseFns[token.EQ] = p.parseInfixExpression
	p.infixParseFns[token.NOT_EQ] = p.parseInfixExpression
	p.infixParseFns[token.LT] = p.parseInfixExpression
	p.infixParseFns[token.GT] = p.parseInfixExpression
	p.infixParseFns[token.LTE] = p.parseInfixExpression
	p.infixParseFns[token.GTE] = p.parseInfixExpression
	p.infixParseFns[token.AND] = p.parseInfixExpression
	p.infixParseFns[token.OR] = p.parseInfixExpression
	p.infixParseFns[token.PLUS_ASSIGN] = p.parseInfixExpression
	p.infixParseFns[token.MINUS_ASSIGN] = p.parseInfixExpression
	p.infixParseFns[token.ASTERISK_ASSIGN] = p.parseInfixExpression
	p.infixParseFns[token.SLASH_ASSIGN] = p.parseInfixExpression
	p.infixParseFns[token.LPAREN] = p.parseCallExpression
	p.infixParseFns[token.LBRACKET] = p.parseIndexExpression
}

// -----------------------------------------------------------------------
// Token navigation
// -----------------------------------------------------------------------

// nextToken advances the two-token window.
// peekToken becomes currentToken and a fresh token is read from the lexer.
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// currentTokenIs returns true if currentToken has the given type.
func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

// peekTokenIs returns true if peekToken has the given type.
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek advances and returns true when peekToken matches the expected type.
// Records a peekError and returns false otherwise.
// This is the primary way the parser enforces grammar rules.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// -----------------------------------------------------------------------
// Entry point
// -----------------------------------------------------------------------

// ParseProgram is the entry point. It builds the root Program node by
// repeatedly parsing statements until EOF.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	for !p.currentTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// -----------------------------------------------------------------------
// Statement parsing
// -----------------------------------------------------------------------

// parseStatement routes to the correct statement parser based on currentToken.
func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement parses: let <name> = <expression> [;]
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.currentToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken() // advance past '='
	stmt.Value = p.parseExpression(LOWEST)

	// optional trailing semicolon
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseReturnStatement parses: return <expression> [;]
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.currentToken}

	p.nextToken() // advance past 'return'
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseWhileStatement parses: while ( <condition> ) { <body> }
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.currentToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken() // advance past '('
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

// parseExpressionStatement wraps a bare expression as a statement.
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.currentToken}
	stmt.Expression = p.parseExpression(LOWEST)

	// semicolons are optional after expression statements
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseBlockStatement parses a { ... } block.
// currentToken must be '{' on entry.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currentToken}
	p.nextToken() // advance past '{'

	for !p.currentTokenIs(token.RBRACE) && !p.currentTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

// -----------------------------------------------------------------------
// Core expression engine — Pratt parsing
// -----------------------------------------------------------------------

// parseExpression is the Pratt parsing loop.
//
//  1. Look up the prefix function for currentToken.
//  2. If none → record error, return nil.
//  3. Call prefix function → left.
//  4. While peekToken is not ';' AND peekPrecedence > precedence:
//     a. Look up infix function for peekToken.
//     b. Advance.
//     c. Call infix function with left → new left.
//  5. Return left.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefixFn := p.prefixParseFns[p.currentToken.Type]
	if prefixFn == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}
	left := prefixFn()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infixFn := p.infixParseFns[p.peekToken.Type]
		if infixFn == nil {
			return left
		}
		p.nextToken()
		left = infixFn(left)
	}
	return left
}

// -----------------------------------------------------------------------
// Prefix parse functions
// -----------------------------------------------------------------------

// parseIdentifier wraps currentToken in an Identifier node.
// No advancement — the Pratt loop handles that.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

// parseIntegerLiteral converts the token literal to int64.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.currentToken}
	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currentToken.Literal)
		p.errors = append(p.errors, NewParseError(p.currentToken, msg))
		return nil
	}
	lit.Value = value
	return lit
}

// parseFloatLiteral converts the token literal to float64.
func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.currentToken}
	value, err := strconv.ParseFloat(p.currentToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.currentToken.Literal)
		p.errors = append(p.errors, NewParseError(p.currentToken, msg))
		return nil
	}
	lit.Value = value
	return lit
}

// parseStringLiteral wraps the token literal in a StringLiteral node.
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}
}

// parseBoolean checks whether currentToken is TRUE or FALSE.
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.currentToken,
		Value: p.currentTokenIs(token.TRUE),
	}
}

// parseNullLiteral wraps the null token in a NullLiteral node.
func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.currentToken}
}

// parsePrefixExpression handles:  !<expr>  -<expr>  ++<expr>  --<expr>
func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
	}
	p.nextToken() // advance past the operator
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

// parseGroupedExpression handles ( <expr> ).
// The grouping itself produces no AST node — it just resets precedence to LOWEST
// inside the parentheses.
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // advance past '('
	expr := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return expr
}

// parseIfExpression handles: if ( <condition> ) { <consequence> } [else { <alternative> }]
func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: p.currentToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken() // advance past '('
	expr.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expr.Consequence = p.parseBlockStatement()

	// optional else branch
	if p.peekTokenIs(token.ELSE) {
		p.nextToken() // advance to 'else'
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		expr.Alternative = p.parseBlockStatement()
	}
	return expr
}

// parseFunctionLiteral handles: fn ( <params> ) { <body> }
// The keyword is "fn" but the token type is token.FUNCTION.
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.currentToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

// parseFunctionParameters parses the comma-separated identifier list between ( and ).
// currentToken must be '(' on entry. Returns an empty slice for zero parameters.
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var identifiers []*ast.Identifier

	// zero parameters: fn() { ... }
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken() // advance to ')'
		return identifiers
	}

	p.nextToken() // advance past '(' to first identifier
	identifiers = append(identifiers, &ast.Identifier{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	})

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // advance to ','
		p.nextToken() // advance to next identifier
		identifiers = append(identifiers, &ast.Identifier{
			Token: p.currentToken,
			Value: p.currentToken.Literal,
		})
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers
}

// parseArrayLiteral handles: [ <expr>, <expr>, ... ]
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.currentToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// parseHashLiteral handles: { <expr> : <expr>, ... }
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{
		Token: p.currentToken,
		Pairs: make(map[ast.Expression]ast.Expression),
	}

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken() // advance to key
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken() // advance past ':'
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hash
}

// -----------------------------------------------------------------------
// Infix parse functions
// -----------------------------------------------------------------------

// parseInfixExpression handles all binary operators.
// left is the already-parsed left operand.
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
		Left:     left,
	}
	precedence := p.currentPrecedence()
	p.nextToken() // advance past the operator
	expr.Right = p.parseExpression(precedence)
	return expr
}

// parseCallExpression handles: <expr> ( <args> )
// left is the function expression.
func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	call := &ast.CallExpression{Token: p.currentToken, Function: left}
	call.Arguments = p.parseExpressionList(token.RPAREN)
	return call
}

// parseIndexExpression handles: <expr> [ <index> ]
// left is the object being indexed.
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	expr := &ast.IndexExpression{Token: p.currentToken, Left: left}
	p.nextToken() // advance past '['
	expr.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return expr
}

// -----------------------------------------------------------------------
// Shared helpers
// -----------------------------------------------------------------------

// parseExpressionList parses a comma-separated list of expressions terminated
// by the given end token. Used for function arguments and array elements.
// currentToken must be the opening delimiter on entry.
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	var list []ast.Expression

	// empty list
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken() // advance past opening delimiter to first element
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // advance to ','
		p.nextToken() // advance to next element
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}
	return list
}
