package object

import (
	"testing"

	"github.com/Vamshi-gande/zenlang/pkg/ast"
	"github.com/Vamshi-gande/zenlang/pkg/token"
)

// ---------------------------------------------------------------------------
// Integer
// ---------------------------------------------------------------------------

func TestIntegerType(t *testing.T) {
	i := &Integer{Value: 42}
	if i.Type() != INTEGER_OBJ {
		t.Errorf("Type() = %q, want %q", i.Type(), INTEGER_OBJ)
	}
}

func TestIntegerInspect(t *testing.T) {
	tests := []struct {
		value int64
		want  string
	}{
		{0, "0"},
		{42, "42"},
		{-7, "-7"},
		{1000000, "1000000"},
	}
	for _, tt := range tests {
		i := &Integer{Value: tt.value}
		if i.Inspect() != tt.want {
			t.Errorf("Integer(%d).Inspect() = %q, want %q", tt.value, i.Inspect(), tt.want)
		}
	}
}

func TestIntegerHashKey(t *testing.T) {
	a := &Integer{Value: 1}
	b := &Integer{Value: 1}
	c := &Integer{Value: 2}

	// Same value → same HashKey
	if a.HashKey() != b.HashKey() {
		t.Error("Integer{1}.HashKey() != Integer{1}.HashKey() — should be equal")
	}
	// Different value → different HashKey
	if a.HashKey() == c.HashKey() {
		t.Error("Integer{1}.HashKey() == Integer{2}.HashKey() — should differ")
	}
}

func TestIntegerHashKeyIncludesType(t *testing.T) {
	// Integer(1) and Boolean(true) both encode value=1 in their raw uint64,
	// but the HashKey.Type field must distinguish them.
	i := &Integer{Value: 1}
	b := TRUE // Boolean{true} also encodes 1
	if i.HashKey() == b.HashKey() {
		t.Error("Integer{1}.HashKey() should not equal Boolean{true}.HashKey()")
	}
}

// ---------------------------------------------------------------------------
// Float
// ---------------------------------------------------------------------------

func TestFloatType(t *testing.T) {
	f := &Float{Value: 3.14}
	if f.Type() != FLOAT_OBJ {
		t.Errorf("Type() = %q, want %q", f.Type(), FLOAT_OBJ)
	}
}

func TestFloatInspect(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{3.14, "3.14"},
		{0.0, "0"},
		{-1.5, "-1.5"},
		{100.0, "100"},
	}
	for _, tt := range tests {
		f := &Float{Value: tt.value}
		if f.Inspect() != tt.want {
			t.Errorf("Float(%v).Inspect() = %q, want %q", tt.value, f.Inspect(), tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Boolean singletons
// ---------------------------------------------------------------------------

func TestBooleanSingletons(t *testing.T) {
	if TRUE.Value != true {
		t.Error("TRUE singleton should have Value=true")
	}
	if FALSE.Value != false {
		t.Error("FALSE singleton should have Value=false")
	}
	// There must only ever be one true and one false instance
	if TRUE == FALSE {
		t.Error("TRUE and FALSE must be distinct pointers")
	}
}

func TestBooleanType(t *testing.T) {
	if TRUE.Type() != BOOLEAN_OBJ {
		t.Errorf("TRUE.Type() = %q, want %q", TRUE.Type(), BOOLEAN_OBJ)
	}
	if FALSE.Type() != BOOLEAN_OBJ {
		t.Errorf("FALSE.Type() = %q, want %q", FALSE.Type(), BOOLEAN_OBJ)
	}
}

func TestBooleanInspect(t *testing.T) {
	if TRUE.Inspect() != "true" {
		t.Errorf("TRUE.Inspect() = %q, want \"true\"", TRUE.Inspect())
	}
	if FALSE.Inspect() != "false" {
		t.Errorf("FALSE.Inspect() = %q, want \"false\"", FALSE.Inspect())
	}
}

func TestBooleanHashKey(t *testing.T) {
	// Same value → same HashKey
	if TRUE.HashKey() != TRUE.HashKey() {
		t.Error("TRUE.HashKey() should be stable")
	}
	if FALSE.HashKey() != FALSE.HashKey() {
		t.Error("FALSE.HashKey() should be stable")
	}
	// Different values → different HashKeys
	if TRUE.HashKey() == FALSE.HashKey() {
		t.Error("TRUE.HashKey() should not equal FALSE.HashKey()")
	}
}

// ---------------------------------------------------------------------------
// Null singleton
// ---------------------------------------------------------------------------

func TestNullSingleton(t *testing.T) {
	if NULL == nil {
		t.Fatal("NULL singleton must not be a Go nil")
	}
	if NULL.Type() != NULL_OBJ {
		t.Errorf("NULL.Type() = %q, want %q", NULL.Type(), NULL_OBJ)
	}
	if NULL.Inspect() != "null" {
		t.Errorf("NULL.Inspect() = %q, want \"null\"", NULL.Inspect())
	}
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func TestStringType(t *testing.T) {
	s := &String{Value: "hello"}
	if s.Type() != STRING_OBJ {
		t.Errorf("Type() = %q, want %q", s.Type(), STRING_OBJ)
	}
}

func TestStringInspect(t *testing.T) {
	s := &String{Value: "hello world"}
	if s.Inspect() != "hello world" {
		t.Errorf("Inspect() = %q, want \"hello world\"", s.Inspect())
	}
}

func TestStringHashKey(t *testing.T) {
	a := &String{Value: "hello"}
	b := &String{Value: "hello"}
	c := &String{Value: "world"}

	// Same content → same HashKey
	if a.HashKey() != b.HashKey() {
		t.Error("String{hello}.HashKey() != String{hello}.HashKey() — should be equal")
	}
	// Different content → different HashKey
	if a.HashKey() == c.HashKey() {
		t.Error("String{hello}.HashKey() == String{world}.HashKey() — should differ")
	}
}

func TestStringHashKeyIncludesType(t *testing.T) {
	// A string whose content happens to hash the same as an integer should
	// still differ because HashKey.Type is different.
	s := &String{Value: "test"}
	i := &Integer{Value: int64(s.HashKey().Value)} // manufacture a collision candidate
	if s.HashKey() == i.HashKey() {
		t.Error("String and Integer HashKeys should differ due to Type field")
	}
}

// ---------------------------------------------------------------------------
// ReturnValue
// ---------------------------------------------------------------------------

func TestReturnValueType(t *testing.T) {
	rv := &ReturnValue{Value: &Integer{Value: 5}}
	if rv.Type() != RETURN_VALUE_OBJ {
		t.Errorf("Type() = %q, want %q", rv.Type(), RETURN_VALUE_OBJ)
	}
}

func TestReturnValueInspect(t *testing.T) {
	rv := &ReturnValue{Value: &Integer{Value: 42}}
	if rv.Inspect() != "42" {
		t.Errorf("Inspect() = %q, want \"42\"", rv.Inspect())
	}
}

func TestReturnValueWrapsNull(t *testing.T) {
	rv := &ReturnValue{Value: NULL}
	if rv.Inspect() != "null" {
		t.Errorf("Inspect() = %q, want \"null\"", rv.Inspect())
	}
}

// ---------------------------------------------------------------------------
// Error
// ---------------------------------------------------------------------------

func TestErrorType(t *testing.T) {
	e := &Error{Message: "something went wrong"}
	if e.Type() != ERROR_OBJ {
		t.Errorf("Type() = %q, want %q", e.Type(), ERROR_OBJ)
	}
}

func TestErrorInspect(t *testing.T) {
	e := &Error{Message: "identifier not found: x"}
	want := "ERROR: identifier not found: x"
	if e.Inspect() != want {
		t.Errorf("Inspect() = %q, want %q", e.Inspect(), want)
	}
}

func TestErrorInspectEmptyMessage(t *testing.T) {
	e := &Error{Message: ""}
	if e.Inspect() != "ERROR: " {
		t.Errorf("Inspect() = %q, want \"ERROR: \"", e.Inspect())
	}
}

// ---------------------------------------------------------------------------
// Builtin
// ---------------------------------------------------------------------------

func TestBuiltinType(t *testing.T) {
	b := &Builtin{Fn: func(args ...Object) Object { return NULL }}
	if b.Type() != BUILTIN_OBJ {
		t.Errorf("Type() = %q, want %q", b.Type(), BUILTIN_OBJ)
	}
}

func TestBuiltinInspect(t *testing.T) {
	b := &Builtin{Fn: func(args ...Object) Object { return NULL }}
	if b.Inspect() != "builtin function" {
		t.Errorf("Inspect() = %q, want \"builtin function\"", b.Inspect())
	}
}

func TestBuiltinFnIsCallable(t *testing.T) {
	called := false
	b := &Builtin{Fn: func(args ...Object) Object {
		called = true
		return NULL
	}}
	b.Fn()
	if !called {
		t.Error("Builtin.Fn should be callable")
	}
}

func TestBuiltinFnReceivesArgs(t *testing.T) {
	var received []Object
	b := &Builtin{Fn: func(args ...Object) Object {
		received = args
		return NULL
	}}
	arg1 := &Integer{Value: 1}
	arg2 := &String{Value: "hi"}
	b.Fn(arg1, arg2)
	if len(received) != 2 {
		t.Fatalf("expected 2 args, got %d", len(received))
	}
	if received[0] != arg1 || received[1] != arg2 {
		t.Error("Builtin.Fn received wrong arguments")
	}
}

// ---------------------------------------------------------------------------
// Array
// ---------------------------------------------------------------------------

func TestArrayType(t *testing.T) {
	a := &Array{Elements: []Object{&Integer{Value: 1}}}
	if a.Type() != ARRAY_OBJ {
		t.Errorf("Type() = %q, want %q", a.Type(), ARRAY_OBJ)
	}
}

func TestArrayInspectEmpty(t *testing.T) {
	a := &Array{Elements: []Object{}}
	if a.Inspect() != "[]" {
		t.Errorf("Inspect() = %q, want \"[]\"", a.Inspect())
	}
}

func TestArrayInspectThreeElements(t *testing.T) {
	a := &Array{Elements: []Object{
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}
	want := "[1, 2, 3]"
	if a.Inspect() != want {
		t.Errorf("Inspect() = %q, want %q", a.Inspect(), want)
	}
}

func TestArrayInspectMixedTypes(t *testing.T) {
	a := &Array{Elements: []Object{
		&Integer{Value: 1},
		TRUE,
		&String{Value: "hello"},
	}}
	want := "[1, true, hello]"
	if a.Inspect() != want {
		t.Errorf("Inspect() = %q, want %q", a.Inspect(), want)
	}
}

// ---------------------------------------------------------------------------
// Hash
// ---------------------------------------------------------------------------

func TestHashType(t *testing.T) {
	h := &Hash{Pairs: make(map[HashKey]HashPair)}
	if h.Type() != HASH_OBJ {
		t.Errorf("Type() = %q, want %q", h.Type(), HASH_OBJ)
	}
}

func TestHashInspectEmpty(t *testing.T) {
	h := &Hash{Pairs: make(map[HashKey]HashPair)}
	if h.Inspect() != "{}" {
		t.Errorf("Inspect() = %q, want \"{}\"", h.Inspect())
	}
}

func TestHashInspectOneEntry(t *testing.T) {
	key := &String{Value: "name"}
	val := &String{Value: "Alice"}
	h := &Hash{
		Pairs: map[HashKey]HashPair{
			key.HashKey(): {Key: key, Value: val},
		},
	}
	result := h.Inspect()
	// Can't assert exact order but must contain key and value
	if !strContains(result, "name") {
		t.Errorf("Inspect() missing key, got %q", result)
	}
	if !strContains(result, "Alice") {
		t.Errorf("Inspect() missing value, got %q", result)
	}
	if !strContains(result, ":") {
		t.Errorf("Inspect() missing colon separator, got %q", result)
	}
}

func TestHashKeyEqualityAcrossTypes(t *testing.T) {
	// All three hashable types should never collide with each other for the
	// same uint64 value because HashKey.Type distinguishes them.
	intKey := (&Integer{Value: 1}).HashKey()
	boolKey := TRUE.HashKey() // also encodes 1
	strKey := (&String{Value: "1"}).HashKey()

	if intKey == boolKey {
		t.Error("Integer and Boolean HashKeys with value 1 should differ")
	}
	if intKey == strKey {
		t.Error("Integer and String HashKeys should differ")
	}
	if boolKey == strKey {
		t.Error("Boolean and String HashKeys should differ")
	}
}

// ---------------------------------------------------------------------------
// Function
// ---------------------------------------------------------------------------

func TestFunctionType(t *testing.T) {
	env := NewEnvironment()
	fn := &Function{
		Parameters: []*ast.Identifier{},
		Body: &ast.BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
		},
		Env: env,
	}
	if fn.Type() != FUNCTION_OBJ {
		t.Errorf("Type() = %q, want %q", fn.Type(), FUNCTION_OBJ)
	}
}

func TestFunctionInspectNoParams(t *testing.T) {
	env := NewEnvironment()
	fn := &Function{
		Parameters: []*ast.Identifier{},
		Body: &ast.BlockStatement{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: nil,
		},
		Env: env,
	}
	result := fn.Inspect()
	if !strContains(result, "fn") {
		t.Errorf("Inspect() missing 'fn', got %q", result)
	}
	if !strContains(result, "()") {
		t.Errorf("Inspect() missing '()', got %q", result)
	}
}

func TestFunctionInspectWithParams(t *testing.T) {
	env := NewEnvironment()
	fn := &Function{
		Parameters: []*ast.Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
			{Token: token.Token{Type: token.IDENT, Literal: "y"}, Value: "y"},
		},
		Body: &ast.BlockStatement{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: nil,
		},
		Env: env,
	}
	result := fn.Inspect()
	if !strContains(result, "x") || !strContains(result, "y") {
		t.Errorf("Inspect() missing parameters, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// Environment — basic Get/Set
// ---------------------------------------------------------------------------

func TestEnvironmentSetAndGet(t *testing.T) {
	env := NewEnvironment()
	env.Set("x", &Integer{Value: 5})

	obj, ok := env.Get("x")
	if !ok {
		t.Fatal("Get('x') returned ok=false after Set")
	}
	i, ok := obj.(*Integer)
	if !ok {
		t.Fatalf("Get('x') returned %T, want *Integer", obj)
	}
	if i.Value != 5 {
		t.Errorf("Get('x').Value = %d, want 5", i.Value)
	}
}

func TestEnvironmentGetMissing(t *testing.T) {
	env := NewEnvironment()
	_, ok := env.Get("undefined")
	if ok {
		t.Error("Get on undefined name should return ok=false")
	}
}

func TestEnvironmentSetReturnsValue(t *testing.T) {
	env := NewEnvironment()
	val := &Integer{Value: 42}
	returned := env.Set("x", val)
	if returned != val {
		t.Error("Set should return the stored object")
	}
}

func TestEnvironmentOverwrite(t *testing.T) {
	env := NewEnvironment()
	env.Set("x", &Integer{Value: 1})
	env.Set("x", &Integer{Value: 99})

	obj, _ := env.Get("x")
	if obj.(*Integer).Value != 99 {
		t.Errorf("overwritten value should be 99, got %d", obj.(*Integer).Value)
	}
}

// ---------------------------------------------------------------------------
// Environment — lexical scoping
// ---------------------------------------------------------------------------

func TestEnclosedEnvironmentSeesOuterScope(t *testing.T) {
	outer := NewEnvironment()
	outer.Set("x", &Integer{Value: 10})

	inner := NewEnclosedEnvironment(outer)
	obj, ok := inner.Get("x")
	if !ok {
		t.Fatal("enclosed env should see variable from outer scope")
	}
	if obj.(*Integer).Value != 10 {
		t.Errorf("x from outer = %d, want 10", obj.(*Integer).Value)
	}
}

func TestEnclosedEnvironmentDoesNotLeakToOuter(t *testing.T) {
	outer := NewEnvironment()
	inner := NewEnclosedEnvironment(outer)
	inner.Set("y", &Integer{Value: 7})

	_, ok := outer.Get("y")
	if ok {
		t.Error("variable set in inner scope should not be visible in outer scope")
	}
}

func TestEnclosedEnvironmentShadowsOuterVariable(t *testing.T) {
	outer := NewEnvironment()
	outer.Set("x", &Integer{Value: 1})

	inner := NewEnclosedEnvironment(outer)
	inner.Set("x", &Integer{Value: 99}) // shadows outer x

	// Inner should see its own x
	obj, _ := inner.Get("x")
	if obj.(*Integer).Value != 99 {
		t.Errorf("inner x = %d, want 99 (should shadow outer)", obj.(*Integer).Value)
	}
	// Outer x should be unchanged
	obj, _ = outer.Get("x")
	if obj.(*Integer).Value != 1 {
		t.Errorf("outer x = %d, want 1 (should not be affected by inner)", obj.(*Integer).Value)
	}
}

func TestThreeLevelScopeChain(t *testing.T) {
	// global → function → block
	global := NewEnvironment()
	global.Set("g", &Integer{Value: 1})

	funcEnv := NewEnclosedEnvironment(global)
	funcEnv.Set("f", &Integer{Value: 2})

	blockEnv := NewEnclosedEnvironment(funcEnv)
	blockEnv.Set("b", &Integer{Value: 3})

	// blockEnv can see all three
	for name, want := range map[string]int64{"g": 1, "f": 2, "b": 3} {
		obj, ok := blockEnv.Get(name)
		if !ok {
			t.Errorf("blockEnv.Get(%q) returned ok=false", name)
			continue
		}
		if obj.(*Integer).Value != want {
			t.Errorf("blockEnv.Get(%q) = %d, want %d", name, obj.(*Integer).Value, want)
		}
	}

	// global can only see g
	if _, ok := global.Get("f"); ok {
		t.Error("global should not see 'f' from funcEnv")
	}
	if _, ok := global.Get("b"); ok {
		t.Error("global should not see 'b' from blockEnv")
	}
}

func TestClosureEnvironmentSurvivesOuterReturn(t *testing.T) {
	// Simulates a closure: the inner function captures the outer env.
	// Even after "outer returns" (we discard outerCallEnv from our local scope),
	// the closure's reference keeps it alive.
	outerCallEnv := NewEnclosedEnvironment(NewEnvironment())
	outerCallEnv.Set("count", &Integer{Value: 0})

	// Simulate returning a closure that points to outerCallEnv
	closureEnv := outerCallEnv

	// "outer function returns" — we no longer hold outerCallEnv directly
	// but closureEnv still references it
	outerCallEnv = nil
	_ = outerCallEnv // suppress unused warning

	obj, ok := closureEnv.Get("count")
	if !ok {
		t.Fatal("closure should still see captured variable after outer scope would have ended")
	}
	if obj.(*Integer).Value != 0 {
		t.Errorf("count = %d, want 0", obj.(*Integer).Value)
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func strContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
