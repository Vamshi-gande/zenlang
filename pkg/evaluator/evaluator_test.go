package evaluator

import (
	"testing"

	"github.com/Vamshi-gande/zenlang/pkg/lexer"
	"github.com/Vamshi-gande/zenlang/pkg/object"
	"github.com/Vamshi-gande/zenlang/pkg/parser"
)

// ---------------------------------------------------------------------------
// Core test helper
// ---------------------------------------------------------------------------

// testEval drives the full pipeline: source string → lexer → parser → Eval.
// Every test in this file goes through this function.
func testEval(input string) object.Object {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	return Eval(program, env)
}

// assertInteger asserts that obj is an *object.Integer with the expected value.
func assertInteger(t *testing.T, obj object.Object, expected int64) {
	t.Helper()
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Fatalf("expected *object.Integer, got %T (%+v)", obj, obj)
	}
	if result.Value != expected {
		t.Errorf("Integer value: got %d, want %d", result.Value, expected)
	}
}

// assertFloat asserts that obj is an *object.Float with the expected value.
func assertFloat(t *testing.T, obj object.Object, expected float64) {
	t.Helper()
	result, ok := obj.(*object.Float)
	if !ok {
		t.Fatalf("expected *object.Float, got %T (%+v)", obj, obj)
	}
	if result.Value != expected {
		t.Errorf("Float value: got %f, want %f", result.Value, expected)
	}
}

// assertBoolean asserts that obj is the TRUE or FALSE singleton.
func assertBoolean(t *testing.T, obj object.Object, expected bool) {
	t.Helper()
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Fatalf("expected *object.Boolean, got %T (%+v)", obj, obj)
	}
	if result.Value != expected {
		t.Errorf("Boolean value: got %t, want %t", result.Value, expected)
	}
}

// assertNull asserts that obj is the NULL singleton.
func assertNull(t *testing.T, obj object.Object) {
	t.Helper()
	if obj != object.NULL {
		t.Errorf("expected NULL, got %T (%+v)", obj, obj)
	}
}

// assertError asserts that obj is an *object.Error and that its message
// contains the expected substring.
func assertError(t *testing.T, obj object.Object, expectedMsg string) {
	t.Helper()
	errObj, ok := obj.(*object.Error)
	if !ok {
		t.Fatalf("expected *object.Error, got %T (%+v)", obj, obj)
	}
	if !contains(errObj.Message, expectedMsg) {
		t.Errorf("Error message: got %q, want it to contain %q", errObj.Message, expectedMsg)
	}
}

// assertString asserts that obj is an *object.String with the expected value.
func assertString(t *testing.T, obj object.Object, expected string) {
	t.Helper()
	result, ok := obj.(*object.String)
	if !ok {
		t.Fatalf("expected *object.String, got %T (%+v)", obj, obj)
	}
	if result.Value != expected {
		t.Errorf("String value: got %q, want %q", result.Value, expected)
	}
}

// contains is a minimal substring helper used in assertError.
func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Integer arithmetic
// ---------------------------------------------------------------------------

func TestIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5", 10},
		{"5 - 5", 0},
		{"5 * 5", 25},
		{"10 / 2", 5},
		{"5 + 5 + 5 - 5", 10},
		{"2 * 3 + 4", 10},
		{"2 + 3 * 4", 14},
		{"5 * 5 + 10", 35},
		{"2 * (5 + 10)", 30},
		{"(2 + 3) * (4 - 1)", 15},
		{"50 / 2 * 2 + 10", 60},
		{"-50 + 100 + -50", 0},
	}
	for _, tt := range tests {
		assertInteger(t, testEval(tt.input), tt.expected)
	}
}

func TestIntegerDivisionByZero(t *testing.T) {
	assertError(t, testEval("10 / 0"), "division by zero")
}

// ---------------------------------------------------------------------------
// Float arithmetic
// ---------------------------------------------------------------------------

func TestFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"1.5 + 1.5", 3.0},
		{"10.0 / 4.0", 2.5},
		{"2 * 1.5", 3.0}, // int * float promotion
		{"1.5 + 1", 2.5}, // float + int promotion
	}
	for _, tt := range tests {
		assertFloat(t, testEval(tt.input), tt.expected)
	}
}

func TestFloatComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.5 < 2.0", true},
		{"2.0 > 1.5", true},
		{"1.5 == 1.5", true},
		{"1.5 != 2.0", true},
		{"1.0 >= 1.0", true},
		{"1.0 <= 0.9", false},
	}
	for _, tt := range tests {
		assertBoolean(t, testEval(tt.input), tt.expected)
	}
}

// ---------------------------------------------------------------------------
// Boolean expressions
// ---------------------------------------------------------------------------

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"1 <= 1", true},
		{"1 >= 1", true},
		{"2 <= 1", false},
		{"1 >= 2", false},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
	}
	for _, tt := range tests {
		assertBoolean(t, testEval(tt.input), tt.expected)
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false && false", false},
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
		{"true || true", true},
	}
	for _, tt := range tests {
		assertBoolean(t, testEval(tt.input), tt.expected)
	}
}

// ---------------------------------------------------------------------------
// Bang operator
// ---------------------------------------------------------------------------

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"!5", false}, // non-null non-false values are truthy
		{"!!5", true},
		{"!null", true}, // null is falsy
		{"!!null", false},
	}
	for _, tt := range tests {
		assertBoolean(t, testEval(tt.input), tt.expected)
	}
}

// ---------------------------------------------------------------------------
// Minus prefix
// ---------------------------------------------------------------------------

func TestMinusPrefix(t *testing.T) {
	assertInteger(t, testEval("-5"), -5)
	assertInteger(t, testEval("-10"), -10)
	assertInteger(t, testEval("- -5"), 5) // double negation: two separate minus tokens
	assertError(t, testEval("-true"), "unknown operator")
}

// ---------------------------------------------------------------------------
// If / else expressions
// ---------------------------------------------------------------------------

func TestIfElseExpression(t *testing.T) {
	assertInteger(t, testEval("if (true) { 10 }"), 10)
	assertInteger(t, testEval("if (1 < 2) { 10 }"), 10)
	assertInteger(t, testEval("if (1 < 2) { 10 } else { 20 }"), 10)
	assertInteger(t, testEval("if (1 > 2) { 10 } else { 20 }"), 20)
	assertNull(t, testEval("if (false) { 10 }"))
	assertNull(t, testEval("if (1 > 2) { 10 }"))
}

func TestIfElseWithExpressionCondition(t *testing.T) {
	assertInteger(t, testEval("if (5 == 5) { 42 } else { 0 }"), 42)
	assertInteger(t, testEval("if (5 != 5) { 42 } else { 99 }"), 99)
}

// ---------------------------------------------------------------------------
// Return statements
// ---------------------------------------------------------------------------

func TestReturnStatements(t *testing.T) {
	assertInteger(t, testEval("return 10;"), 10)
	assertInteger(t, testEval("return 10; 9;"), 10)
	assertInteger(t, testEval("9; return 10; 9;"), 10)
	assertInteger(t, testEval("return 2 * 5; 9;"), 10)
}

func TestReturnInsideNestedBlocks(t *testing.T) {
	// return inside a nested if block must still propagate all the way out
	input := `
if (true) {
    if (true) {
        return 10;
    }
    return 1;
}
`
	assertInteger(t, testEval(input), 10)
}

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		expectedMsg string
	}{
		{"5 + true", "type mismatch"},
		{"5 + true; 5", "type mismatch"}, // error stops execution
		{"-true", "unknown operator"},
		{"true + false", "unknown operator"},
		{"5; true + false; 5", "unknown operator"},
		{"foobar", "identifier not found"},
		{`{"name": "Alice"}[fn(x){x}]`, "unusable as hash key"},
	}
	for _, tt := range tests {
		assertError(t, testEval(tt.input), tt.expectedMsg)
	}
}

func TestErrorStopsExecution(t *testing.T) {
	// The second expression should never run
	result := testEval("5 + true; 100")
	assertError(t, result, "type mismatch")
}

// ---------------------------------------------------------------------------
// Null literal
// ---------------------------------------------------------------------------

func TestNullLiteral(t *testing.T) {
	assertNull(t, testEval("null"))
}

// ---------------------------------------------------------------------------
// Let statements and variables
// ---------------------------------------------------------------------------

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let x = 5; x", 5},
		{"let x = 5 * 5; x", 25},
		{"let x = 5; let y = x; y", 5},
		{"let x = 5; let y = x + 5; y", 10},
		{"let x = 5; let y = x; let z = x + y + 5; z", 15},
	}
	for _, tt := range tests {
		assertInteger(t, testEval(tt.input), tt.expected)
	}
}

func TestUndefinedVariable(t *testing.T) {
	assertError(t, testEval("foobar"), "identifier not found")
}

// ---------------------------------------------------------------------------
// Compound assignment operators
// ---------------------------------------------------------------------------

func TestCompoundAssignment(t *testing.T) {
	assertInteger(t, testEval("let x = 5; x += 3; x"), 8)
	assertInteger(t, testEval("let x = 10; x -= 4; x"), 6)
	assertInteger(t, testEval("let x = 3; x *= 4; x"), 12)
	assertInteger(t, testEval("let x = 20; x /= 4; x"), 5)
}

// ---------------------------------------------------------------------------
// Increment and decrement
// ---------------------------------------------------------------------------

func TestIncrementDecrement(t *testing.T) {
	assertInteger(t, testEval("let x = 5; ++x; x"), 6)
	assertInteger(t, testEval("let x = 5; --x; x"), 4)
	assertInteger(t, testEval("let x = 0; ++x"), 1)
}

// ---------------------------------------------------------------------------
// String expressions
// ---------------------------------------------------------------------------

func TestStringLiteral(t *testing.T) {
	assertString(t, testEval(`"hello"`), "hello")
	assertString(t, testEval(`"hello world"`), "hello world")
}

func TestStringConcatenation(t *testing.T) {
	assertString(t, testEval(`"hello" + " " + "world"`), "hello world")
	assertString(t, testEval(`"foo" + "bar"`), "foobar")
}

func TestStringEquality(t *testing.T) {
	assertBoolean(t, testEval(`"hello" == "hello"`), true)
	assertBoolean(t, testEval(`"hello" == "world"`), false)
	assertBoolean(t, testEval(`"hello" != "world"`), true)
}

func TestStringOperatorError(t *testing.T) {
	assertError(t, testEval(`"hello" - "world"`), "unknown operator")
}

// ---------------------------------------------------------------------------
// Functions
// ---------------------------------------------------------------------------

func TestFunctionLiteral(t *testing.T) {
	result := testEval("fn(x) { x + 2 }")
	fn, ok := result.(*object.Function)
	if !ok {
		t.Fatalf("expected *object.Function, got %T", result)
	}
	if len(fn.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
	}
	if fn.Parameters[0].Value != "x" {
		t.Errorf("parameter name: got %q, want %q", fn.Parameters[0].Value, "x")
	}
}

func TestFunctionCall(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x }; identity(5)", 5},
		{"let identity = fn(x) { return x; }; identity(5)", 5},
		{"let double = fn(x) { x * 2 }; double(5)", 10},
		{"let add = fn(a, b) { a + b }; add(5, 5)", 10},
		{"let add = fn(a, b) { a + b }; add(5 + 5, add(3, 3))", 16},
		{"fn(x) { x }(5)", 5},
		{"fn(x, y) { x * y }(3, 4)", 12},
	}
	for _, tt := range tests {
		assertInteger(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionCallNoArgs(t *testing.T) {
	assertInteger(t, testEval("let f = fn() { 42 }; f()"), 42)
}

// ---------------------------------------------------------------------------
// Closures
// ---------------------------------------------------------------------------

func TestClosure(t *testing.T) {
	input := `
let newAdder = fn(x) {
    fn(y) { x + y }
}
let addTwo = newAdder(2)
addTwo(3)
`
	assertInteger(t, testEval(input), 5)
}

func TestMakeCounter(t *testing.T) {
	input := `
let makeCounter = fn() {
    let count = 0
    fn() {
        count = count + 1
        count
    }
}
let counter = makeCounter()
counter()
counter()
counter()
`
	// The last call returns 3 — count was incremented three times
	// Note: this tests that the closure mutates the captured environment
	// via re-binding (count = count + 1 uses compound let re-assignment)
	assertInteger(t, testEval(input), 3)
}

func TestClosureCapturesOuterVariable(t *testing.T) {
	input := `
let x = 10
let addX = fn(n) { n + x }
addX(5)
`
	assertInteger(t, testEval(input), 15)
}

// ---------------------------------------------------------------------------
// Recursion
// ---------------------------------------------------------------------------

func TestFactorial(t *testing.T) {
	input := `
let factorial = fn(n) {
    if (n <= 1) { return 1 }
    return n * factorial(n - 1)
}
factorial(5)
`
	assertInteger(t, testEval(input), 120)
}

func TestFactorialZero(t *testing.T) {
	input := `
let factorial = fn(n) {
    if (n <= 1) { return 1 }
    return n * factorial(n - 1)
}
factorial(0)
`
	assertInteger(t, testEval(input), 1)
}

func TestFibonacci(t *testing.T) {
	input := `
let fib = fn(n) {
    if (n <= 0) { return 0 }
    if (n == 1) { return 1 }
    return fib(n - 1) + fib(n - 2)
}
fib(10)
`
	assertInteger(t, testEval(input), 55)
}

// ---------------------------------------------------------------------------
// While loops
// ---------------------------------------------------------------------------

func TestWhileLoop(t *testing.T) {
	input := `
let x = 0
let i = 0
while (i < 5) {
    x = x + 1
    i += 1
}
x
`
	assertInteger(t, testEval(input), 5)
}

func TestWhileLoopNeverExecutes(t *testing.T) {
	input := `
let x = 42
while (false) {
    x = 0
}
x
`
	assertInteger(t, testEval(input), 42)
}

func TestWhileLoopReturnInsideBody(t *testing.T) {
	input := `
let findFirst = fn(arr, target) {
    let i = 0
    while (i < 3) {
        if (arr[i] == target) { return i }
        i += 1
    }
    return -1
}
findFirst([10, 20, 30], 20)
`
	assertInteger(t, testEval(input), 1)
}

// ---------------------------------------------------------------------------
// Arrays
// ---------------------------------------------------------------------------

func TestArrayLiteral(t *testing.T) {
	result := testEval("[1, 2, 3]")
	arr, ok := result.(*object.Array)
	if !ok {
		t.Fatalf("expected *object.Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
	assertInteger(t, arr.Elements[0], 1)
	assertInteger(t, arr.Elements[1], 2)
	assertInteger(t, arr.Elements[2], 3)
}

func TestArrayIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"[1, 2, 3][0]", 1},
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][2]", 3},
		{"let i = 0; [1][i]", 1},
		{"[1, 2, 3][1 + 1]", 3},
	}
	for _, tt := range tests {
		assertInteger(t, testEval(tt.input), tt.expected)
	}
}

func TestArrayIndexOutOfBounds(t *testing.T) {
	assertNull(t, testEval("[1, 2, 3][3]"))
	assertNull(t, testEval("[1, 2, 3][-1]"))
	assertNull(t, testEval("[][0]"))
}

func TestArrayEmptyLiteral(t *testing.T) {
	result := testEval("[]")
	arr, ok := result.(*object.Array)
	if !ok {
		t.Fatalf("expected *object.Array, got %T", result)
	}
	if len(arr.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(arr.Elements))
	}
}

// ---------------------------------------------------------------------------
// Hash maps
// ---------------------------------------------------------------------------

func TestHashLiteralStringKeys(t *testing.T) {
	input := `
let h = {"one": 1, "two": 2, "three": 3}
h["one"]
`
	assertInteger(t, testEval(input), 1)
}

func TestHashLiteralIntegerKeys(t *testing.T) {
	// integer key 1 maps to the string "one"
	result := testEval(`{1: "one"}[1]`)
	assertString(t, result, "one")
	// integer key 2 maps to the string "two"
	result2 := testEval(`{1: "one", 2: "two"}[2]`)
	assertString(t, result2, "two")
}

func TestHashLiteralBooleanKeys(t *testing.T) {
	assertInteger(t, testEval(`{true: 1, false: 0}[true]`), 1)
	assertInteger(t, testEval(`{true: 1, false: 0}[false]`), 0)
}

func TestHashMissingKey(t *testing.T) {
	assertNull(t, testEval(`{"one": 1}["two"]`))
	assertNull(t, testEval(`{}["missing"]`))
}

func TestHashUnusableKeyError(t *testing.T) {
	assertError(t, testEval(`{"name": "Alice"}[fn(x){x}]`), "unusable as hash key")
	assertError(t, testEval(`{"name": "Alice"}[[1,2,3]]`), "unusable as hash key")
}

// ---------------------------------------------------------------------------
// Builtin: len
// ---------------------------------------------------------------------------

func TestBuiltinLen(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`len("")`, 0},
		{`len("hello")`, 5},
		{`len("hello world")`, 11},
		{`len([])`, 0},
		{`len([1, 2, 3])`, 3},
		{`len([1, 2, 3, 4, 5])`, 5},
	}
	for _, tt := range tests {
		assertInteger(t, testEval(tt.input), tt.expected)
	}
}

func TestBuiltinLenErrors(t *testing.T) {
	assertError(t, testEval(`len(1)`), "argument to `len` not supported")
	assertError(t, testEval(`len(true)`), "argument to `len` not supported")
	assertError(t, testEval(`len("one", "two")`), "wrong number of arguments")
}

// ---------------------------------------------------------------------------
// Builtin: first
// ---------------------------------------------------------------------------

func TestBuiltinFirst(t *testing.T) {
	assertInteger(t, testEval(`first([1, 2, 3])`), 1)
	assertInteger(t, testEval(`first([42])`), 42)
	assertNull(t, testEval(`first([])`))
}

func TestBuiltinFirstErrors(t *testing.T) {
	assertError(t, testEval(`first(1)`), "argument to `first` must be ARRAY")
	assertError(t, testEval(`first([1], [2])`), "wrong number of arguments")
}

// ---------------------------------------------------------------------------
// Builtin: last
// ---------------------------------------------------------------------------

func TestBuiltinLast(t *testing.T) {
	assertInteger(t, testEval(`last([1, 2, 3])`), 3)
	assertInteger(t, testEval(`last([42])`), 42)
	assertNull(t, testEval(`last([])`))
}

func TestBuiltinLastErrors(t *testing.T) {
	assertError(t, testEval(`last(1)`), "argument to `last` must be ARRAY")
	assertError(t, testEval(`last([1], [2])`), "wrong number of arguments")
}

// ---------------------------------------------------------------------------
// Builtin: push
// ---------------------------------------------------------------------------

func TestBuiltinPush(t *testing.T) {
	result := testEval(`push([1, 2], 3)`)
	arr, ok := result.(*object.Array)
	if !ok {
		t.Fatalf("expected *object.Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
	assertInteger(t, arr.Elements[2], 3)
}

func TestBuiltinPushDoesNotMutateOriginal(t *testing.T) {
	input := `
let a = [1, 2, 3]
let b = push(a, 4)
len(a)
`
	// original array must still have 3 elements
	assertInteger(t, testEval(input), 3)
}

func TestBuiltinPushOntoEmpty(t *testing.T) {
	result := testEval(`push([], 1)`)
	arr, ok := result.(*object.Array)
	if !ok {
		t.Fatalf("expected *object.Array, got %T", result)
	}
	assertInteger(t, arr.Elements[0], 1)
}

func TestBuiltinPushErrors(t *testing.T) {
	assertError(t, testEval(`push(1, 2)`), "first argument to `push` must be ARRAY")
	assertError(t, testEval(`push([1])`), "wrong number of arguments")
}

// ---------------------------------------------------------------------------
// Builtin: type
// ---------------------------------------------------------------------------

func TestBuiltinType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`type(5)`, "INTEGER"},
		{`type(3.14)`, "FLOAT"},
		{`type("hello")`, "STRING"},
		{`type(true)`, "BOOLEAN"},
		{`type(false)`, "BOOLEAN"},
		{`type(null)`, "NULL"},
		{`type([1,2,3])`, "ARRAY"},
		{`type({"a": 1})`, "HASH"},
		{`type(fn(x){x})`, "FUNCTION"},
	}
	for _, tt := range tests {
		assertString(t, testEval(tt.input), tt.expected)
	}
}

func TestBuiltinTypeErrors(t *testing.T) {
	assertError(t, testEval(`type()`), "wrong number of arguments")
	assertError(t, testEval(`type(1, 2)`), "wrong number of arguments")
}

// ---------------------------------------------------------------------------
// Calling non-functions
// ---------------------------------------------------------------------------

func TestCallingNonFunction(t *testing.T) {
	assertError(t, testEval("let x = 5; x()"), "not a function")
	assertError(t, testEval(`"hello"()`), "not a function")
}

// ---------------------------------------------------------------------------
// Truthiness
// ---------------------------------------------------------------------------

func TestTruthiness(t *testing.T) {
	// truthy: non-null, non-false
	assertInteger(t, testEval("if (1) { 1 } else { 0 }"), 1)
	assertInteger(t, testEval(`if ("") { 1 } else { 0 }`), 1) // empty string is truthy
	assertInteger(t, testEval("if (0) { 1 } else { 0 }"), 1)  // 0 is truthy
	// falsy: null and false
	assertInteger(t, testEval("if (false) { 1 } else { 0 }"), 0)
	assertInteger(t, testEval("if (null) { 1 } else { 0 }"), 0)
}

// ---------------------------------------------------------------------------
// Full programs
// ---------------------------------------------------------------------------

func TestSumArray(t *testing.T) {
	input := `
let sum = fn(arr) {
    let total = 0
    let i = 0
    while (i < len(arr)) {
        total = total + arr[i]
        i += 1
    }
    total
}
sum([1, 2, 3, 4, 5])
`
	assertInteger(t, testEval(input), 15)
}

func TestMapFunction(t *testing.T) {
	input := `
let map = fn(arr, f) {
    let result = []
    let i = 0
    while (i < len(arr)) {
        result = push(result, f(arr[i]))
        i += 1
    }
    result
}
let doubled = map([1, 2, 3], fn(x) { x * 2 })
doubled[2]
`
	assertInteger(t, testEval(input), 6)
}

func TestHigherOrderFunctions(t *testing.T) {
	input := `
let apply = fn(f, x) { f(x) }
let triple = fn(x) { x * 3 }
apply(triple, 5)
`
	assertInteger(t, testEval(input), 15)
}

func TestFunctionReturningFunction(t *testing.T) {
	input := `
let multiplier = fn(factor) {
    fn(x) { x * factor }
}
let double = multiplier(2)
let triple = multiplier(3)
double(5) + triple(4)
`
	assertInteger(t, testEval(input), 22) // 10 + 12
}
