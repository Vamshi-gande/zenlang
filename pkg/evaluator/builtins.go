package evaluator

import (
	"fmt"

	"github.com/Vamshi-gande/zenlang/pkg/object"
)

// builtins is the package-level map of built-in function names to Builtin
// objects. evalIdentifier checks this map after failing to find a name in
// the current environment, so every entry here is always in scope.
var builtins = map[string]*object.Builtin{
	"len":   {Fn: lenBuiltin},
	"print": {Fn: printBuiltin},
	"first": {Fn: firstBuiltin},
	"last":  {Fn: lastBuiltin},
	"push":  {Fn: pushBuiltin},
	"type":  {Fn: typeBuiltin},
}

// ---------------------------------------------------------------------------
// len
// ---------------------------------------------------------------------------

// lenBuiltin returns the length of a string or array.
// Errors on wrong number of arguments or unsupported type.
//
//	len("hello")  → 5
//	len([1,2,3])  → 3
//	len(5)        → ERROR: argument to `len` not supported, got INTEGER
var lenBuiltin = func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments to `len`: got %d, want 1", len(args))
	}
	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}

// ---------------------------------------------------------------------------
// print
// ---------------------------------------------------------------------------

// printBuiltin writes all arguments to stdout, space-separated, followed by
// a newline. Always returns NULL — print is used for its side effect only.
//
//	print("hello")        → (prints "hello\n")   returns null
//	print(1, true, "hi")  → (prints "1 true hi\n") returns null
var printBuiltin = func(args ...object.Object) object.Object {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = arg.Inspect()
	}
	output := ""
	for i, p := range parts {
		if i > 0 {
			output += " "
		}
		output += p
	}
	fmt.Println(output)
	return object.NULL
}

// ---------------------------------------------------------------------------
// first
// ---------------------------------------------------------------------------

// firstBuiltin returns the first element of an array, or NULL if empty.
// Errors on wrong number of arguments or non-array argument.
//
//	first([1,2,3])  → 1
//	first([])       → null
//	first(5)        → ERROR: argument to `first` must be ARRAY, got INTEGER
var firstBuiltin = func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments to `first`: got %d, want 1", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
	}
	if len(arr.Elements) == 0 {
		return object.NULL
	}
	return arr.Elements[0]
}

// ---------------------------------------------------------------------------
// last
// ---------------------------------------------------------------------------

// lastBuiltin returns the last element of an array, or NULL if empty.
// Errors on wrong number of arguments or non-array argument.
//
//	last([1,2,3])  → 3
//	last([])       → null
//	last("hi")     → ERROR: argument to `last` must be ARRAY, got STRING
var lastBuiltin = func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments to `last`: got %d, want 1", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
	}
	length := len(arr.Elements)
	if length == 0 {
		return object.NULL
	}
	return arr.Elements[length-1]
}

// ---------------------------------------------------------------------------
// push
// ---------------------------------------------------------------------------

// pushBuiltin returns a NEW array with the given element appended to the end.
// The original array is never mutated — Zen arrays are immutable through this
// interface.
//
//	push([1,2,3], 4)  → [1, 2, 3, 4]
//	push([], 1)       → [1]
//	push(5, 1)        → ERROR: first argument to `push` must be ARRAY, got INTEGER
var pushBuiltin = func(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments to `push`: got %d, want 2", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("first argument to `push` must be ARRAY, got %s", args[0].Type())
	}
	newElements := make([]object.Object, len(arr.Elements)+1)
	copy(newElements, arr.Elements)
	newElements[len(arr.Elements)] = args[1]
	return &object.Array{Elements: newElements}
}

// ---------------------------------------------------------------------------
// type
// ---------------------------------------------------------------------------

// typeBuiltin returns a STRING describing the Zen type of its argument.
//
//	type(5)       → "INTEGER"
//	type("hello") → "STRING"
//	type(true)    → "BOOLEAN"
//	type(null)    → "NULL"
//	type([1,2])   → "ARRAY"
var typeBuiltin = func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments to `type`: got %d, want 1", len(args))
	}
	return &object.String{Value: string(args[0].Type())}
}
