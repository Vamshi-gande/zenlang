package ast

import (
	"testing"

	"github.com/Vamshi-gande/zenlang/pkg/token"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func strContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Program
// ---------------------------------------------------------------------------

func TestProgramTokenLiteralEmpty(t *testing.T) {
	p := &Program{}
	if p.TokenLiteral() != "" {
		t.Errorf("empty Program.TokenLiteral() = %q, want \"\"", p.TokenLiteral())
	}
}

func TestProgramTokenLiteralNonEmpty(t *testing.T) {
	p := &Program{
		Statements: []Statement{
			&ExpressionStatement{
				Token: token.Token{Type: token.INT, Literal: "42"},
				Expression: &IntegerLiteral{
					Token: token.Token{Type: token.INT, Literal: "42"},
					Value: 42,
				},
			},
		},
	}
	if p.TokenLiteral() != "42" {
		t.Errorf("Program.TokenLiteral() = %q, want \"42\"", p.TokenLiteral())
	}
}

func TestProgramStringConcatenatesStatements(t *testing.T) {
	p := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
				Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
			},
			&ReturnStatement{
				Token:       token.Token{Type: token.RETURN, Literal: "return"},
				ReturnValue: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
			},
		},
	}
	result := p.String()
	if !strContains(result, "let x = 5;") {
		t.Errorf("expected 'let x = 5;' in %q", result)
	}
	if !strContains(result, "return x;") {
		t.Errorf("expected 'return x;' in %q", result)
	}
}

// ---------------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------------

func TestLetStatementString(t *testing.T) {
	stmt := &LetStatement{
		Token: token.Token{Type: token.LET, Literal: "let"},
		Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Value: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "anotherVar"}, Value: "anotherVar"},
	}
	want := "let x = anotherVar;"
	if stmt.String() != want {
		t.Errorf("LetStatement.String() = %q, want %q", stmt.String(), want)
	}
}

func TestLetStatementNilValue(t *testing.T) {
	// Value can be nil while the parser is partially built — should not panic.
	stmt := &LetStatement{
		Token: token.Token{Type: token.LET, Literal: "let"},
		Name:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Value: nil,
	}
	want := "let x = ;"
	if stmt.String() != want {
		t.Errorf("LetStatement(nil value).String() = %q, want %q", stmt.String(), want)
	}
}

func TestReturnStatementString(t *testing.T) {
	stmt := &ReturnStatement{
		Token:       token.Token{Type: token.RETURN, Literal: "return"},
		ReturnValue: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
	}
	want := "return x;"
	if stmt.String() != want {
		t.Errorf("ReturnStatement.String() = %q, want %q", stmt.String(), want)
	}
}

func TestExpressionStatementString(t *testing.T) {
	stmt := &ExpressionStatement{
		Token: token.Token{Type: token.INT, Literal: "5"},
		Expression: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "5"},
			Value: 5,
		},
	}
	if stmt.String() != "5" {
		t.Errorf("ExpressionStatement.String() = %q, want \"5\"", stmt.String())
	}
}

func TestExpressionStatementNilExpression(t *testing.T) {
	stmt := &ExpressionStatement{
		Token:      token.Token{Type: token.INT, Literal: "5"},
		Expression: nil,
	}
	if stmt.String() != "" {
		t.Errorf("ExpressionStatement(nil).String() = %q, want \"\"", stmt.String())
	}
}

func TestBlockStatementString(t *testing.T) {
	block := &BlockStatement{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Statements: []Statement{
			&ExpressionStatement{
				Token:      token.Token{Type: token.IDENT, Literal: "x"},
				Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
			},
		},
	}
	if block.String() != "x" {
		t.Errorf("BlockStatement.String() = %q, want \"x\"", block.String())
	}
}

func TestWhileStatementString(t *testing.T) {
	// while x { x }
	ws := &WhileStatement{
		Token:     token.Token{Type: token.WHILE, Literal: "while"},
		Condition: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Body: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token:      token.Token{Type: token.IDENT, Literal: "x"},
					Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
				},
			},
		},
	}
	result := ws.String()
	if !strContains(result, "while") {
		t.Errorf("WhileStatement.String() missing 'while', got %q", result)
	}
}

// ---------------------------------------------------------------------------
// Expressions — literals
// ---------------------------------------------------------------------------

func TestIntegerLiteralString(t *testing.T) {
	il := &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "42"}, Value: 42}
	if il.String() != "42" {
		t.Errorf("IntegerLiteral.String() = %q, want \"42\"", il.String())
	}
}

func TestFloatLiteralString(t *testing.T) {
	fl := &FloatLiteral{Token: token.Token{Type: token.FLOAT, Literal: "3.14"}, Value: 3.14}
	if fl.String() != "3.14" {
		t.Errorf("FloatLiteral.String() = %q, want \"3.14\"", fl.String())
	}
}

func TestStringLiteralString(t *testing.T) {
	sl := &StringLiteral{Token: token.Token{Type: token.STRING, Literal: "hello"}, Value: "hello"}
	if sl.String() != "hello" {
		t.Errorf("StringLiteral.String() = %q, want \"hello\"", sl.String())
	}
}

func TestBooleanLiteralTrue(t *testing.T) {
	bl := &BooleanLiteral{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true}
	if bl.String() != "true" {
		t.Errorf("BooleanLiteral(true).String() = %q, want \"true\"", bl.String())
	}
}

func TestBooleanLiteralFalse(t *testing.T) {
	bl := &BooleanLiteral{Token: token.Token{Type: token.FALSE, Literal: "false"}, Value: false}
	if bl.String() != "false" {
		t.Errorf("BooleanLiteral(false).String() = %q, want \"false\"", bl.String())
	}
}

func TestNullLiteralString(t *testing.T) {
	nl := &NullLiteral{Token: token.Token{Type: token.NULL, Literal: "null"}}
	if nl.String() != "null" {
		t.Errorf("NullLiteral.String() = %q, want \"null\"", nl.String())
	}
}

func TestIdentifierString(t *testing.T) {
	id := &Identifier{Token: token.Token{Type: token.IDENT, Literal: "myVar"}, Value: "myVar"}
	if id.String() != "myVar" {
		t.Errorf("Identifier.String() = %q, want \"myVar\"", id.String())
	}
}

// ---------------------------------------------------------------------------
// Expressions — operators
// ---------------------------------------------------------------------------

func TestPrefixExpressionBang(t *testing.T) {
	// (!true)
	expr := &PrefixExpression{
		Token:    token.Token{Type: token.BANG, Literal: "!"},
		Operator: "!",
		Right:    &BooleanLiteral{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
	}
	if expr.String() != "(!true)" {
		t.Errorf("PrefixExpression.String() = %q, want \"(!true)\"", expr.String())
	}
}

func TestPrefixExpressionMinus(t *testing.T) {
	// (-5)
	expr := &PrefixExpression{
		Token:    token.Token{Type: token.MINUS, Literal: "-"},
		Operator: "-",
		Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
	}
	if expr.String() != "(-5)" {
		t.Errorf("PrefixExpression.String() = %q, want \"(-5)\"", expr.String())
	}
}

func TestPrefixExpressionIncrement(t *testing.T) {
	// (++x)  — token.INC = "++"
	expr := &PrefixExpression{
		Token:    token.Token{Type: token.INC, Literal: "++"},
		Operator: "++",
		Right:    &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
	}
	if expr.String() != "(++x)" {
		t.Errorf("PrefixExpression(++).String() = %q, want \"(++x)\"", expr.String())
	}
}

func TestInfixExpressionPlus(t *testing.T) {
	// (5 + 3)
	expr := &InfixExpression{
		Token:    token.Token{Type: token.PLUS, Literal: "+"},
		Operator: "+",
		Left:     &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
		Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "3"}, Value: 3},
	}
	if expr.String() != "(5 + 3)" {
		t.Errorf("InfixExpression.String() = %q, want \"(5 + 3)\"", expr.String())
	}
}

func TestInfixExpressionEquality(t *testing.T) {
	// (x == y)  — token.EQ = "=="
	expr := &InfixExpression{
		Token:    token.Token{Type: token.EQ, Literal: "=="},
		Operator: "==",
		Left:     &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Right:    &Identifier{Token: token.Token{Type: token.IDENT, Literal: "y"}, Value: "y"},
	}
	if expr.String() != "(x == y)" {
		t.Errorf("InfixExpression(==).String() = %q, want \"(x == y)\"", expr.String())
	}
}

func TestInfixExpressionAnd(t *testing.T) {
	// (a && b)  — token.AND = "&&"
	expr := &InfixExpression{
		Token:    token.Token{Type: token.AND, Literal: "&&"},
		Operator: "&&",
		Left:     &Identifier{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
		Right:    &Identifier{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
	}
	if expr.String() != "(a && b)" {
		t.Errorf("InfixExpression(&&).String() = %q, want \"(a && b)\"", expr.String())
	}
}

func TestInfixExpressionPlusAssign(t *testing.T) {
	// (x += 1)  — token.PLUS_ASSIGN = "+="
	expr := &InfixExpression{
		Token:    token.Token{Type: token.PLUS_ASSIGN, Literal: "+="},
		Operator: "+=",
		Left:     &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
	}
	if expr.String() != "(x += 1)" {
		t.Errorf("InfixExpression(+=).String() = %q, want \"(x += 1)\"", expr.String())
	}
}

// ---------------------------------------------------------------------------
// Expressions — control flow & functions
// ---------------------------------------------------------------------------

func TestIfExpressionNoElse(t *testing.T) {
	expr := &IfExpression{
		Token:     token.Token{Type: token.IF, Literal: "if"},
		Condition: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Consequence: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token:      token.Token{Type: token.IDENT, Literal: "x"},
					Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
				},
			},
		},
		Alternative: nil,
	}
	result := expr.String()
	if !strContains(result, "if") {
		t.Errorf("IfExpression.String() missing 'if', got %q", result)
	}
	if strContains(result, "else") {
		t.Errorf("IfExpression (no else) should not contain 'else', got %q", result)
	}
}

func TestIfExpressionWithElse(t *testing.T) {
	makeBlock := func(name string) *BlockStatement {
		return &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token:      token.Token{Type: token.IDENT, Literal: name},
					Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: name}, Value: name},
				},
			},
		}
	}
	expr := &IfExpression{
		Token:       token.Token{Type: token.IF, Literal: "if"},
		Condition:   &BooleanLiteral{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
		Consequence: makeBlock("x"),
		Alternative: makeBlock("y"),
	}
	result := expr.String()
	if !strContains(result, "else") {
		t.Errorf("IfExpression(else).String() missing 'else', got %q", result)
	}
}

func TestFunctionLiteralString(t *testing.T) {
	// fn(a, b){}
	// Note: keyword "fn" maps to token.FUNCTION in your keywords map
	fn := &FunctionLiteral{
		Token: token.Token{Type: token.FUNCTION, Literal: "fn"},
		Parameters: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
		},
		Body: &BlockStatement{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
	}
	result := fn.String()
	if !strContains(result, "fn") {
		t.Errorf("FunctionLiteral.String() missing 'fn', got %q", result)
	}
	if !strContains(result, "a") || !strContains(result, "b") {
		t.Errorf("FunctionLiteral.String() missing params, got %q", result)
	}
}

func TestCallExpressionString(t *testing.T) {
	// add(1, 2)
	call := &CallExpression{
		Token: token.Token{Type: token.LPAREN, Literal: "("},
		Function: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "add"},
			Value: "add",
		},
		Arguments: []Expression{
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2},
		},
	}
	if call.String() != "add(1, 2)" {
		t.Errorf("CallExpression.String() = %q, want \"add(1, 2)\"", call.String())
	}
}

// ---------------------------------------------------------------------------
// Expressions — collections
// ---------------------------------------------------------------------------

func TestArrayLiteralString(t *testing.T) {
	arr := &ArrayLiteral{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Elements: []Expression{
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "3"}, Value: 3},
		},
	}
	if arr.String() != "[1, 2, 3]" {
		t.Errorf("ArrayLiteral.String() = %q, want \"[1, 2, 3]\"", arr.String())
	}
}

func TestIndexExpressionString(t *testing.T) {
	// (arr[0])
	expr := &IndexExpression{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Left:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "arr"}, Value: "arr"},
		Index: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "0"}, Value: 0},
	}
	if expr.String() != "(arr[0])" {
		t.Errorf("IndexExpression.String() = %q, want \"(arr[0])\"", expr.String())
	}
}

func TestHashLiteralString(t *testing.T) {
	// {"name": "Alice"}
	// COLON token is defined in your token package
	hash := &HashLiteral{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Pairs: map[Expression]Expression{
			&StringLiteral{Token: token.Token{Type: token.STRING, Literal: "name"}, Value: "name"}: &StringLiteral{
				Token: token.Token{Type: token.STRING, Literal: "Alice"},
				Value: "Alice",
			},
		},
	}
	result := hash.String()
	if !strContains(result, "{") || !strContains(result, "}") {
		t.Errorf("HashLiteral.String() missing braces, got %q", result)
	}
	if !strContains(result, ":") {
		t.Errorf("HashLiteral.String() missing colon, got %q", result)
	}
}
