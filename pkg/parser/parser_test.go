package parser

import (
	"fmt"
	"testing"

	"github.com/Vamshi-gande/zenlang/pkg/ast"
	"github.com/Vamshi-gande/zenlang/pkg/lexer"
	"github.com/Vamshi-gande/zenlang/pkg/token"
)

// -----------------------------------------------------------------------
// Test helpers
// -----------------------------------------------------------------------

// parse is a convenience wrapper that lexes and parses an input string,
// fails the test immediately if any parse errors were found, and returns
// the resulting program.
func parse(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	return program
}

// parseWithErrors parses input and returns both the program and the parser
// so the test can inspect errors directly.
func parseWithErrors(input string) (*ast.Program, *Parser) {
	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	return program, p
}

// checkParserErrors fails the test immediately if the parser collected any errors.
func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()
	if len(p.Errors()) == 0 {
		return
	}
	t.Errorf("parser produced %d error(s):", len(p.Errors()))
	for _, e := range p.Errors() {
		t.Errorf("  %s", e.Error())
	}
	t.FailNow()
}

// requireStatements asserts the program has exactly n statements and returns them.
func requireStatements(t *testing.T, program *ast.Program, n int) []ast.Statement {
	t.Helper()
	if len(program.Statements) != n {
		t.Fatalf("program has %d statement(s), want %d", len(program.Statements), n)
	}
	return program.Statements
}

// requireExpressionStatement asserts stmt is an *ast.ExpressionStatement
// and returns the inner expression.
func requireExpressionStatement(t *testing.T, stmt ast.Statement) ast.Expression {
	t.Helper()
	es, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is %T, want *ast.ExpressionStatement", stmt)
	}
	return es.Expression
}

// assertIntegerLiteral checks that expr is an IntegerLiteral with the given value.
func assertIntegerLiteral(t *testing.T, expr ast.Expression, want int64) {
	t.Helper()
	il, ok := expr.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.IntegerLiteral", expr)
	}
	if il.Value != want {
		t.Errorf("IntegerLiteral.Value = %d, want %d", il.Value, want)
	}
}

// assertIdentifier checks that expr is an Identifier with the given name.
func assertIdentifier(t *testing.T, expr ast.Expression, want string) {
	t.Helper()
	id, ok := expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expr is %T, want *ast.Identifier", expr)
	}
	if id.Value != want {
		t.Errorf("Identifier.Value = %q, want %q", id.Value, want)
	}
}

// assertBoolean checks that expr is a BooleanLiteral with the given value.
func assertBoolean(t *testing.T, expr ast.Expression, want bool) {
	t.Helper()
	bl, ok := expr.(*ast.BooleanLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.BooleanLiteral", expr)
	}
	if bl.Value != want {
		t.Errorf("BooleanLiteral.Value = %v, want %v", bl.Value, want)
	}
}

// assertLiteralExpression dispatches to the typed assertion based on the
// concrete type of want.
func assertLiteralExpression(t *testing.T, expr ast.Expression, want interface{}) {
	t.Helper()
	switch v := want.(type) {
	case int:
		assertIntegerLiteral(t, expr, int64(v))
	case int64:
		assertIntegerLiteral(t, expr, v)
	case string:
		assertIdentifier(t, expr, v)
	case bool:
		assertBoolean(t, expr, v)
	default:
		t.Fatalf("assertLiteralExpression: unsupported type %T", v)
	}
}

// assertInfixExpression checks an InfixExpression node fully.
func assertInfixExpression(t *testing.T, expr ast.Expression, left interface{}, operator string, right interface{}) {
	t.Helper()
	ie, ok := expr.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expr is %T, want *ast.InfixExpression", expr)
	}
	assertLiteralExpression(t, ie.Left, left)
	if ie.Operator != operator {
		t.Errorf("InfixExpression.Operator = %q, want %q", ie.Operator, operator)
	}
	assertLiteralExpression(t, ie.Right, right)
}

// -----------------------------------------------------------------------
// Let statements
// -----------------------------------------------------------------------

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input     string
		wantName  string
		wantValue interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foo = y;", "foo", "y"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parse(t, tt.input)
			stmts := requireStatements(t, program, 1)

			ls, ok := stmts[0].(*ast.LetStatement)
			if !ok {
				t.Fatalf("stmt is %T, want *ast.LetStatement", stmts[0])
			}
			if ls.Name.Value != tt.wantName {
				t.Errorf("Name.Value = %q, want %q", ls.Name.Value, tt.wantName)
			}
			assertLiteralExpression(t, ls.Value, tt.wantValue)
		})
	}
}

func TestLetStatementTokenLiteral(t *testing.T) {
	program := parse(t, "let x = 5;")
	ls := program.Statements[0].(*ast.LetStatement)
	if ls.TokenLiteral() != "let" {
		t.Errorf("TokenLiteral = %q, want \"let\"", ls.TokenLiteral())
	}
}

// -----------------------------------------------------------------------
// Return statements
// -----------------------------------------------------------------------

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"return 5;"},
		{"return true;"},
		{"return foobar;"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parse(t, tt.input)
			stmts := requireStatements(t, program, 1)

			rs, ok := stmts[0].(*ast.ReturnStatement)
			if !ok {
				t.Fatalf("stmt is %T, want *ast.ReturnStatement", stmts[0])
			}
			if rs.TokenLiteral() != "return" {
				t.Errorf("TokenLiteral = %q, want \"return\"", rs.TokenLiteral())
			}
			if rs.ReturnValue == nil {
				t.Fatal("ReturnValue is nil")
			}
		})
	}
}

func TestReturnInfixExpression(t *testing.T) {
	program := parse(t, "return x + y;")
	stmts := requireStatements(t, program, 1)
	rs := stmts[0].(*ast.ReturnStatement)
	assertInfixExpression(t, rs.ReturnValue, "x", "+", "y")
}

// -----------------------------------------------------------------------
// Identifier expressions
// -----------------------------------------------------------------------

func TestIdentifierExpression(t *testing.T) {
	program := parse(t, "foobar;")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])
	assertIdentifier(t, expr, "foobar")
}

// -----------------------------------------------------------------------
// Literal expressions
// -----------------------------------------------------------------------

func TestIntegerLiteralExpression(t *testing.T) {
	program := parse(t, "5;")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])
	assertIntegerLiteral(t, expr, 5)
}

func TestFloatLiteralExpression(t *testing.T) {
	program := parse(t, "3.14;")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])
	fl, ok := expr.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.FloatLiteral", expr)
	}
	if fl.Value != 3.14 {
		t.Errorf("FloatLiteral.Value = %v, want 3.14", fl.Value)
	}
}

func TestStringLiteralExpression(t *testing.T) {
	program := parse(t, `"hello"`)
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])
	sl, ok := expr.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.StringLiteral", expr)
	}
	if sl.Value != "hello" {
		t.Errorf("StringLiteral.Value = %q, want \"hello\"", sl.Value)
	}
}

func TestBooleanLiteralTrue(t *testing.T) {
	program := parse(t, "true;")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])
	assertBoolean(t, expr, true)
}

func TestBooleanLiteralFalse(t *testing.T) {
	program := parse(t, "false;")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])
	assertBoolean(t, expr, false)
}

func TestNullLiteral(t *testing.T) {
	program := parse(t, "null;")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])
	if _, ok := expr.(*ast.NullLiteral); !ok {
		t.Fatalf("expr is %T, want *ast.NullLiteral", expr)
	}
}

// -----------------------------------------------------------------------
// Prefix expressions
// -----------------------------------------------------------------------

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		right    interface{}
	}{
		{"!true", "!", true},
		{"!false", "!", false},
		{"-5", "-", 5},
		{"--x", "--", "x"},
		{"++x", "++", "x"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parse(t, tt.input)
			stmts := requireStatements(t, program, 1)
			expr := requireExpressionStatement(t, stmts[0])

			pe, ok := expr.(*ast.PrefixExpression)
			if !ok {
				t.Fatalf("expr is %T, want *ast.PrefixExpression", expr)
			}
			if pe.Operator != tt.operator {
				t.Errorf("Operator = %q, want %q", pe.Operator, tt.operator)
			}
			assertLiteralExpression(t, pe.Right, tt.right)
		})
	}
}

// -----------------------------------------------------------------------
// Infix expressions
// -----------------------------------------------------------------------

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5 + 3", 5, "+", 3},
		{"5 - 3", 5, "-", 3},
		{"5 * 3", 5, "*", 3},
		{"5 / 3", 5, "/", 3},
		{"5 < 3", 5, "<", 3},
		{"5 > 3", 5, ">", 3},
		{"5 <= 3", 5, "<=", 3},
		{"5 >= 3", 5, ">=", 3},
		{"5 == 5", 5, "==", 5},
		{"5 != 3", 5, "!=", 3},
		{"a && b", "a", "&&", "b"},
		{"a || b", "a", "||", "b"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parse(t, tt.input)
			stmts := requireStatements(t, program, 1)
			expr := requireExpressionStatement(t, stmts[0])
			assertInfixExpression(t, expr, tt.left, tt.operator, tt.right)
		})
	}
}

func TestCompoundAssignmentOperators(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"x += 1", "+="},
		{"x -= 1", "-="},
		{"x *= 2", "*="},
		{"x /= 2", "/="},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parse(t, tt.input)
			stmts := requireStatements(t, program, 1)
			expr := requireExpressionStatement(t, stmts[0])
			ie, ok := expr.(*ast.InfixExpression)
			if !ok {
				t.Fatalf("expr is %T, want *ast.InfixExpression", expr)
			}
			if ie.Operator != tt.operator {
				t.Errorf("Operator = %q, want %q", ie.Operator, tt.operator)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Operator precedence — verified via String() output
// -----------------------------------------------------------------------

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"5 + 3 * 2", "(5 + (3 * 2))"},
		{"5 * 3 + 2", "((5 * 3) + 2)"},
		{"(5 + 3) * 2", "((5 + 3) * 2)"},
		{"a + b * c + d", "((a + (b * c)) + d)"},
		{"-a * b", "((-a) * b)"},
		{"!true == false", "((!true) == false)"},
		{"5 > 3 == true", "((5 > 3) == true)"},
		{"1 + 2 + 3", "((1 + 2) + 3)"},
		{"a + b + c * d", "((a + b) + (c * d))"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parse(t, tt.input)
			got := program.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Grouped expressions
// -----------------------------------------------------------------------

func TestGroupedExpression(t *testing.T) {
	program := parse(t, "(5 + 3)")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])
	assertInfixExpression(t, expr, 5, "+", 3)
}

// -----------------------------------------------------------------------
// If expressions
// -----------------------------------------------------------------------

func TestIfExpressionNoElse(t *testing.T) {
	program := parse(t, "if (x < y) { x }")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	ie, ok := expr.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expr is %T, want *ast.IfExpression", expr)
	}
	assertInfixExpression(t, ie.Condition, "x", "<", "y")

	if len(ie.Consequence.Statements) != 1 {
		t.Fatalf("consequence has %d statements, want 1", len(ie.Consequence.Statements))
	}
	assertIdentifier(t, requireExpressionStatement(t, ie.Consequence.Statements[0]), "x")

	if ie.Alternative != nil {
		t.Error("Alternative should be nil for if without else")
	}
}

func TestIfExpressionWithElse(t *testing.T) {
	program := parse(t, "if (x < y) { x } else { y }")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	ie, ok := expr.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expr is %T, want *ast.IfExpression", expr)
	}
	assertInfixExpression(t, ie.Condition, "x", "<", "y")

	if len(ie.Consequence.Statements) != 1 {
		t.Fatalf("consequence has %d statements, want 1", len(ie.Consequence.Statements))
	}
	if ie.Alternative == nil {
		t.Fatal("Alternative is nil, want else block")
	}
	if len(ie.Alternative.Statements) != 1 {
		t.Fatalf("alternative has %d statements, want 1", len(ie.Alternative.Statements))
	}
	assertIdentifier(t, requireExpressionStatement(t, ie.Alternative.Statements[0]), "y")
}

// -----------------------------------------------------------------------
// While statements
// -----------------------------------------------------------------------

func TestWhileStatement(t *testing.T) {
	program := parse(t, "while (x > 0) { x }")
	stmts := requireStatements(t, program, 1)

	ws, ok := stmts[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("stmt is %T, want *ast.WhileStatement", stmts[0])
	}
	assertInfixExpression(t, ws.Condition, "x", ">", 0)

	if len(ws.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(ws.Body.Statements))
	}
	assertIdentifier(t, requireExpressionStatement(t, ws.Body.Statements[0]), "x")
}

// -----------------------------------------------------------------------
// Function literals
// -----------------------------------------------------------------------

func TestFunctionLiteralNoParams(t *testing.T) {
	program := parse(t, "fn() { 5 }")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	fl, ok := expr.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.FunctionLiteral", expr)
	}
	if len(fl.Parameters) != 0 {
		t.Errorf("Parameters count = %d, want 0", len(fl.Parameters))
	}
	if len(fl.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(fl.Body.Statements))
	}
}

func TestFunctionLiteralTwoParams(t *testing.T) {
	program := parse(t, "fn(x, y) { x + y }")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	fl, ok := expr.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.FunctionLiteral", expr)
	}
	if len(fl.Parameters) != 2 {
		t.Fatalf("Parameters count = %d, want 2", len(fl.Parameters))
	}
	assertIdentifier(t, fl.Parameters[0], "x")
	assertIdentifier(t, fl.Parameters[1], "y")

	if len(fl.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(fl.Body.Statements))
	}
	assertInfixExpression(t, requireExpressionStatement(t, fl.Body.Statements[0]), "x", "+", "y")
}

func TestFunctionLiteralThreeParams(t *testing.T) {
	program := parse(t, "fn(a, b, c) {}")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	fl := expr.(*ast.FunctionLiteral)
	if len(fl.Parameters) != 3 {
		t.Fatalf("Parameters count = %d, want 3", len(fl.Parameters))
	}
	assertIdentifier(t, fl.Parameters[0], "a")
	assertIdentifier(t, fl.Parameters[1], "b")
	assertIdentifier(t, fl.Parameters[2], "c")
}

func TestFunctionLiteralTokenLiteral(t *testing.T) {
	// The keyword in source is "fn" which maps to token.FUNCTION — literal stays "fn"
	program := parse(t, "fn() {}")
	fl := requireExpressionStatement(t, program.Statements[0]).(*ast.FunctionLiteral)
	if fl.TokenLiteral() != "fn" {
		t.Errorf("TokenLiteral = %q, want \"fn\"", fl.TokenLiteral())
	}
}

// -----------------------------------------------------------------------
// Call expressions
// -----------------------------------------------------------------------

func TestCallExpressionNoArgs(t *testing.T) {
	program := parse(t, "add()")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	ce, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expr is %T, want *ast.CallExpression", expr)
	}
	assertIdentifier(t, ce.Function, "add")
	if len(ce.Arguments) != 0 {
		t.Errorf("Arguments count = %d, want 0", len(ce.Arguments))
	}
}

func TestCallExpressionTwoArgs(t *testing.T) {
	program := parse(t, "add(1, 2 * 3)")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	ce, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expr is %T, want *ast.CallExpression", expr)
	}
	assertIdentifier(t, ce.Function, "add")

	if len(ce.Arguments) != 2 {
		t.Fatalf("Arguments count = %d, want 2", len(ce.Arguments))
	}
	assertIntegerLiteral(t, ce.Arguments[0], 1)
	assertInfixExpression(t, ce.Arguments[1], 2, "*", 3)
}

func TestCallExpressionImmediatelyInvokedFunctionLiteral(t *testing.T) {
	program := parse(t, "fn(x) { x * 2 }(5)")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	ce, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expr is %T, want *ast.CallExpression", expr)
	}
	if _, ok := ce.Function.(*ast.FunctionLiteral); !ok {
		t.Fatalf("Function field is %T, want *ast.FunctionLiteral", ce.Function)
	}
	if len(ce.Arguments) != 1 {
		t.Fatalf("Arguments count = %d, want 1", len(ce.Arguments))
	}
	assertIntegerLiteral(t, ce.Arguments[0], 5)
}

// -----------------------------------------------------------------------
// Array literals
// -----------------------------------------------------------------------

func TestArrayLiteralEmpty(t *testing.T) {
	program := parse(t, "[]")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	al, ok := expr.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.ArrayLiteral", expr)
	}
	if len(al.Elements) != 0 {
		t.Errorf("Elements count = %d, want 0", len(al.Elements))
	}
}

func TestArrayLiteralThreeElements(t *testing.T) {
	program := parse(t, "[1, 2 * 2, 3 + 3]")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	al, ok := expr.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.ArrayLiteral", expr)
	}
	if len(al.Elements) != 3 {
		t.Fatalf("Elements count = %d, want 3", len(al.Elements))
	}
	assertIntegerLiteral(t, al.Elements[0], 1)
	assertInfixExpression(t, al.Elements[1], 2, "*", 2)
	assertInfixExpression(t, al.Elements[2], 3, "+", 3)
}

// -----------------------------------------------------------------------
// Index expressions
// -----------------------------------------------------------------------

func TestIndexExpression(t *testing.T) {
	program := parse(t, "arr[0]")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	ie, ok := expr.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expr is %T, want *ast.IndexExpression", expr)
	}
	assertIdentifier(t, ie.Left, "arr")
	assertIntegerLiteral(t, ie.Index, 0)
}

func TestIndexExpressionWithExpression(t *testing.T) {
	program := parse(t, "arr[1 + 2]")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	ie := expr.(*ast.IndexExpression)
	assertIdentifier(t, ie.Left, "arr")
	assertInfixExpression(t, ie.Index, 1, "+", 2)
}

// -----------------------------------------------------------------------
// Hash literals
// -----------------------------------------------------------------------

func TestHashLiteralEmpty(t *testing.T) {
	program := parse(t, "{}")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	hl, ok := expr.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.HashLiteral", expr)
	}
	if len(hl.Pairs) != 0 {
		t.Errorf("Pairs count = %d, want 0", len(hl.Pairs))
	}
}

func TestHashLiteralIntegerKeys(t *testing.T) {
	program := parse(t, "{1: 2, 3: 4}")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	hl, ok := expr.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("expr is %T, want *ast.HashLiteral", expr)
	}
	if len(hl.Pairs) != 2 {
		t.Fatalf("Pairs count = %d, want 2", len(hl.Pairs))
	}
}

func TestHashLiteralBooleanKeys(t *testing.T) {
	program := parse(t, "{true: 1, false: 2}")
	stmts := requireStatements(t, program, 1)
	expr := requireExpressionStatement(t, stmts[0])

	hl := expr.(*ast.HashLiteral)
	if len(hl.Pairs) != 2 {
		t.Fatalf("Pairs count = %d, want 2", len(hl.Pairs))
	}
}

// -----------------------------------------------------------------------
// Multiple statements in one program
// -----------------------------------------------------------------------

func TestMultipleStatements(t *testing.T) {
	input := `
let x = 5;
let y = 10;
x + y
`
	program := parse(t, input)
	requireStatements(t, program, 3)

	if _, ok := program.Statements[0].(*ast.LetStatement); !ok {
		t.Errorf("stmt[0] is %T, want *ast.LetStatement", program.Statements[0])
	}
	if _, ok := program.Statements[1].(*ast.LetStatement); !ok {
		t.Errorf("stmt[1] is %T, want *ast.LetStatement", program.Statements[1])
	}
	assertInfixExpression(t, requireExpressionStatement(t, program.Statements[2]), "x", "+", "y")
}

func TestRecursiveFunction(t *testing.T) {
	input := `
let factorial = fn(n) {
    if (n <= 1) { return 1 }
    return n * factorial(n - 1)
}
`
	program := parse(t, input)
	requireStatements(t, program, 1)

	ls := program.Statements[0].(*ast.LetStatement)
	if ls.Name.Value != "factorial" {
		t.Errorf("Name = %q, want \"factorial\"", ls.Name.Value)
	}
	if _, ok := ls.Value.(*ast.FunctionLiteral); !ok {
		t.Fatalf("Value is %T, want *ast.FunctionLiteral", ls.Value)
	}
}

// -----------------------------------------------------------------------
// Error cases
// -----------------------------------------------------------------------

func TestParserErrorMissingIdentifierInLet(t *testing.T) {
	_, p := parseWithErrors("let = 5;")
	if len(p.Errors()) == 0 {
		t.Error("expected parse errors for 'let = 5;', got none")
	}
}

func TestParserErrorMissingAssignInLet(t *testing.T) {
	_, p := parseWithErrors("let x 5;")
	if len(p.Errors()) == 0 {
		t.Error("expected parse errors for 'let x 5;', got none")
	}
}

func TestParserErrorMessageContent(t *testing.T) {
	_, p := parseWithErrors("let = 5;")
	if len(p.Errors()) == 0 {
		t.Fatal("expected at least one error")
	}
	msg := p.Errors()[0].Error()
	if len(msg) == 0 {
		t.Error("error message is empty")
	}
}

func TestParserErrorsDoNotPanic(t *testing.T) {
	badInputs := []string{
		"let",
		"let x",
		"let x =",
		"return",
		"fn(",
		"fn(x",
		"if",
		"if (x",
		"[1, 2",
	}
	for _, input := range badInputs {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("parser panicked on input %q: %v", input, r)
				}
			}()
			_, p := parseWithErrors(input)
			_ = p.Errors()
		})
	}
}

// -----------------------------------------------------------------------
// ParseError type
// -----------------------------------------------------------------------

func TestParseErrorFormat(t *testing.T) {
	tok := token.Token{Type: token.IDENT, Literal: "foo"}
	err := NewParseError(tok, "some message")
	got := err.Error()
	want := "parse error at 'foo': some message"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// -----------------------------------------------------------------------
// Precedence helpers
// -----------------------------------------------------------------------

func TestPeekAndCurrentPrecedence(t *testing.T) {
	l := lexer.NewLexer("5 + 3")
	p := NewParser(l)

	// After construction: currentToken=5, peekToken=+
	if p.peekPrecedence() != SUM {
		t.Errorf("peekPrecedence = %d, want SUM (%d)", p.peekPrecedence(), SUM)
	}
	p.nextToken() // currentToken=+, peekToken=3
	if p.currentPrecedence() != SUM {
		t.Errorf("currentPrecedence = %d, want SUM (%d)", p.currentPrecedence(), SUM)
	}
}

func TestUnknownTokenHasLowestPrecedence(t *testing.T) {
	l := lexer.NewLexer("x")
	p := NewParser(l)
	// peekToken is EOF which has no entry in the precedence table
	if p.peekPrecedence() != LOWEST {
		t.Errorf("peekPrecedence for unknown token = %d, want LOWEST (%d)", p.peekPrecedence(), LOWEST)
	}
}
