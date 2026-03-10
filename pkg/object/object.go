package object

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/Vamshi-gande/zenlang/pkg/ast"
)

// ObjectType is a named string used to identify the kind of a runtime value.
// Using a string type makes debugging easier — errors show "INTEGER" not "3".
type ObjectType string

const (
	INTEGER_OBJ      ObjectType = "INTEGER"
	FLOAT_OBJ        ObjectType = "FLOAT"
	BOOLEAN_OBJ      ObjectType = "BOOLEAN"
	NULL_OBJ         ObjectType = "NULL"
	STRING_OBJ       ObjectType = "STRING"
	RETURN_VALUE_OBJ ObjectType = "RETURN_VALUE"
	ERROR_OBJ        ObjectType = "ERROR"
	FUNCTION_OBJ     ObjectType = "FUNCTION"
	BUILTIN_OBJ      ObjectType = "BUILTIN"
	ARRAY_OBJ        ObjectType = "ARRAY"
	HASH_OBJ         ObjectType = "HASH"
)

// Object is the interface every runtime value in Zen must satisfy.
// Type() identifies the kind. Inspect() produces a human-readable representation
// used by print() and the REPL.
type Object interface {
	Type() ObjectType
	Inspect() string
}

// ---------------------------------------------------------------------------
// Singletons — shared instances for values that never differ
// ---------------------------------------------------------------------------

// TRUE and FALSE are the only two possible boolean values.
// The evaluator always returns these pointers rather than allocating new ones.
var (
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
)

// NULL is the single null value. Returned whenever something produces nothing —
// e.g. an if-expression with no else branch whose condition is false.
var NULL = &Null{}

// ---------------------------------------------------------------------------
// Integer
// ---------------------------------------------------------------------------

// Integer wraps a Go int64 and represents a whole number value at runtime.
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

// HashKey returns a HashKey for use as a map key in Hash objects.
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

// ---------------------------------------------------------------------------
// Float
// ---------------------------------------------------------------------------

// Float wraps a Go float64 and represents a decimal number value at runtime.
type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }

// ---------------------------------------------------------------------------
// Boolean
// ---------------------------------------------------------------------------

// Boolean wraps a Go bool. Always use the TRUE and FALSE singletons
// rather than constructing new Boolean values.
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

// HashKey returns a HashKey so booleans can be used as hash map keys.
func (b *Boolean) HashKey() HashKey {
	var val uint64
	if b.Value {
		val = 1
	}
	return HashKey{Type: b.Type(), Value: val}
}

// ---------------------------------------------------------------------------
// Null
// ---------------------------------------------------------------------------

// Null represents the absence of a value. Always use the NULL singleton.
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

// String wraps a Go string and represents a text value at runtime.
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

// HashKey hashes the string content with FNV-64a so strings can be map keys.
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// ---------------------------------------------------------------------------
// ReturnValue
// ---------------------------------------------------------------------------

// ReturnValue wraps the object being returned from a function.
// The evaluator uses this as a signal to stop executing statements and
// bubble the value back through nested block evaluations.
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// ---------------------------------------------------------------------------
// Error
// ---------------------------------------------------------------------------

// Error represents a runtime error. Like ReturnValue it propagates upward
// through the evaluator — any produced error short-circuits further evaluation.
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// ---------------------------------------------------------------------------
// Function
// ---------------------------------------------------------------------------

// BuiltinFunction is the Go function signature that all built-in functions share.
// It receives evaluated argument objects and returns a result object.
type BuiltinFunction func(args ...Object) Object

// Function is a first-class value in Zen. It carries the AST nodes for its
// parameters and body, plus the environment that was active when it was defined.
// That captured environment is what makes closures work.
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	params := make([]string, len(f.Parameters))
	for i, p := range f.Parameters {
		params[i] = p.String()
	}
	return fmt.Sprintf("fn(%s) {\n%s\n}", strings.Join(params, ", "), f.Body.String())
}

// ---------------------------------------------------------------------------
// Builtin
// ---------------------------------------------------------------------------

// Builtin wraps a Go function as a Zen value. Used for built-ins like
// len(), print(), push() — functions implemented in Go, not Zen.
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

// ---------------------------------------------------------------------------
// Array
// ---------------------------------------------------------------------------

// Array holds an ordered list of runtime objects.
type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	elements := make([]string, len(a.Elements))
	for i, el := range a.Elements {
		elements[i] = el.Inspect()
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

// ---------------------------------------------------------------------------
// Hash
// ---------------------------------------------------------------------------

// HashKey is a comparable struct used as a Go map key inside Hash objects.
// It pairs the object's type with a uint64 hash of its value, so that
// Integer(1), Boolean(true), and String("1") all produce distinct keys
// even if their hash values happened to collide.
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Hashable must be implemented by any object type that can be used as a
// hash map key. Currently: Integer, Boolean, String.
type Hashable interface {
	HashKey() HashKey
}

// HashPair stores both the original key object and its associated value.
// This lets Inspect() display the original key rather than a raw uint64.
type HashPair struct {
	Key   Object
	Value Object
}

// Hash represents a key-value collection. Keys must implement Hashable.
type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	pairs := make([]string, 0, len(h.Pairs))
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}
