# `pkg/object` — Package Documentation

**Package:** `github.com/Vamshi-gande/zenlang/pkg/object`
**Location:** `pkg/object/`
**Files:** `object.go`, `environment.go`, `object_test.go`

---

## Overview

The object package is the **runtime type system** of Zen. While the AST represents the structure of source code, the object package represents the actual values that exist while the program is running. Every computation the evaluator performs produces an object, every variable holds an object, and every function returns an object.

```
AST Nodes (structure) → Evaluator → Object Values (runtime)
                                           ↑
                                    defined here
```

The package has two distinct responsibilities: defining all runtime value types in `object.go`, and providing scoped variable storage in `environment.go`.

---

## File Structure

```
pkg/object/
├── object.go        → Object interface, all runtime types, HashKey system
├── environment.go   → Environment struct, lexical scope chain
└── object_test.go   → Tests for all types and environment behaviour
```

---

## Dependencies

```
object.go
    ├── pkg/ast        (ast.Identifier, ast.BlockStatement — used in Function)
    ├── fmt            (Inspect() formatting)
    ├── strings        (Inspect() joining)
    └── hash/fnv       (String.HashKey() hashing)

environment.go
    └── Object interface (same package, no external imports)

object_test.go
    ├── pkg/ast        (constructing Function nodes in tests)
    └── pkg/token      (token.Token values for AST node construction)
```

---

## `object.go`

### The Object Interface

Every single value in Zen at runtime implements this interface:

```go
type ObjectType string

type Object interface {
    Type() ObjectType
    Inspect() string
}
```

**`Type()`** returns an `ObjectType` string constant that identifies what kind of value this is. The evaluator checks this constantly — you can't add a function to an integer, and `Type()` is how the evaluator knows which operation is legal.

**`Inspect()`** returns a human-readable string. This is what `print()` and the REPL display to the user.

Using a named string type for `ObjectType` makes debugging readable — error messages show `"INTEGER"` instead of an opaque numeric constant.

### Object Type Constants

```
INTEGER_OBJ      = "INTEGER"
FLOAT_OBJ        = "FLOAT"
BOOLEAN_OBJ      = "BOOLEAN"
NULL_OBJ         = "NULL"
STRING_OBJ       = "STRING"
RETURN_VALUE_OBJ = "RETURN_VALUE"
ERROR_OBJ        = "ERROR"
FUNCTION_OBJ     = "FUNCTION"
BUILTIN_OBJ      = "BUILTIN"
ARRAY_OBJ        = "ARRAY"
HASH_OBJ         = "HASH"
```

### Singletons

Three values are declared as package-level variables. The evaluator always returns these pointers rather than constructing new instances:

```go
var TRUE  = &Boolean{Value: true}
var FALSE = &Boolean{Value: false}
var NULL  = &Null{}
```

For booleans there are only ever two possible values. Allocating a new `Boolean` on every `true` or `false` evaluation wastes memory and breaks pointer equality — the evaluator can test `result == TRUE` as a fast boolean check. The same reasoning applies to `NULL`, of which there is only ever one.

---

## Object Types — Complete Reference

### Integer

Wraps a Go `int64`. Represents all whole number values at runtime.

```go
type Integer struct {
    Value int64
}
```

| Method | Returns |
|---|---|
| `Type()` | `INTEGER_OBJ` |
| `Inspect()` | Decimal string: `"42"`, `"-7"`, `"0"` |
| `HashKey()` | `HashKey{Type: INTEGER_OBJ, Value: uint64(Value)}` |

`Integer` implements `Hashable` — integers can be used as hash map keys.

---

### Float

Wraps a Go `float64`. Represents decimal number values at runtime. Added to complete the type pipeline alongside the lexer's `FLOAT` token and the AST's `FloatLiteral` node.

```go
type Float struct {
    Value float64
}
```

| Method | Returns |
|---|---|
| `Type()` | `FLOAT_OBJ` |
| `Inspect()` | Uses `%g` format: `"3.14"`, `"-1.5"`, `"100"` (trailing zeros stripped), `"0"` |

`Float` does **not** implement `Hashable`. Floating point equality is unreliable in IEEE 754 — `0.1 + 0.2 != 0.3` — so floats make dangerous map keys and are excluded by design.

---

### Boolean

Wraps a Go `bool`. Always use the `TRUE` and `FALSE` singletons — never construct `&Boolean{...}` directly in the evaluator.

```go
type Boolean struct {
    Value bool
}
```

| Method | Returns |
|---|---|
| `Type()` | `BOOLEAN_OBJ` |
| `Inspect()` | `"true"` or `"false"` |
| `HashKey()` | `true` → `HashKey{BOOLEAN_OBJ, 1}`, `false` → `HashKey{BOOLEAN_OBJ, 0}` |

`Boolean` implements `Hashable`.

---

### Null

Represents the absence of a value. Returned when an `if` expression without an `else` branch has a false condition, or when a function body ends without a `return`. Always use the `NULL` singleton.

```go
type Null struct{}
```

| Method | Returns |
|---|---|
| `Type()` | `NULL_OBJ` |
| `Inspect()` | `"null"` (fixed string, no fields) |

`Null` does **not** implement `Hashable`.

---

### String

Wraps a Go `string`. The `Literal` from the lexer's `STRING` token — with surrounding quotes already stripped — becomes this `Value`.

```go
type String struct {
    Value string
}
```

| Method | Returns |
|---|---|
| `Type()` | `STRING_OBJ` |
| `Inspect()` | Returns `Value` directly, no added quotes |
| `HashKey()` | FNV-64a hash of `Value` bytes, paired with `STRING_OBJ` type tag |

`String` implements `Hashable`. Hashing uses `hash/fnv` (`fnv.New64a()`), the standard Go choice for non-cryptographic string hashing.

---

### ReturnValue

A wrapper object that carries another object being returned from a function. The evaluator uses this as an upward propagation signal.

```go
type ReturnValue struct {
    Value Object
}
```

| Method | Returns |
|---|---|
| `Type()` | `RETURN_VALUE_OBJ` |
| `Inspect()` | Delegates to `Value.Inspect()` |

**How it works:** When the evaluator hits a `return` statement it wraps the value in `ReturnValue` and passes it upward. At each enclosing level the evaluator checks: is this a `ReturnValue`? If yes, unwrap and stop. This means `return` works correctly even inside deeply nested `if` blocks or `while` loops — the signal travels all the way back to the function call boundary without any special unwinding logic at each layer.

---

### Error

Represents a runtime error. Like `ReturnValue`, it propagates upward through the evaluator and short-circuits all further evaluation.

```go
type Error struct {
    Message string
}
```

| Method | Returns |
|---|---|
| `Type()` | `ERROR_OBJ` |
| `Inspect()` | `"ERROR: " + Message` — e.g. `"ERROR: identifier not found: x"` |

The evaluator checks for errors after almost every sub-evaluation. As soon as an `Error` is produced it is passed upward immediately, without evaluating anything else in the current expression or statement.

---

### Function

The most important object type. A function value in Zen is first-class — it can be stored in variables, passed as arguments, and returned from other functions. It carries three things from the moment it is created:

```go
type Function struct {
    Parameters []*ast.Identifier
    Body       *ast.BlockStatement
    Env        *Environment
}
```

**`Parameters`** — the identifier list from the function literal in the AST. The evaluator uses these names to bind call arguments into the function's enclosed scope on each invocation.

**`Body`** — the block statement from the function literal. The evaluator re-walks this AST on every call. (Phase 2 will replace this with compiled bytecode.)

**`Env`** — the environment that was active **when the function was defined**, not when it is called. This captured environment is what makes closures work — the function carries its defining scope wherever it goes.

| Method | Returns |
|---|---|
| `Type()` | `FUNCTION_OBJ` |
| `Inspect()` | `fn(x, y) {\n<body>\n}` — reconstructed source form |

`Function` holds actual AST nodes, which is why `object.go` imports `pkg/ast`.

---

### Builtin

Wraps a Go function as a Zen-callable value. Used for `len()`, `print()`, `push()`, and other standard library functions that are implemented in Go rather than Zen.

```go
type BuiltinFunction func(args ...Object) Object

type Builtin struct {
    Fn BuiltinFunction
}
```

`BuiltinFunction` is the signature every built-in must match — variadic `Object` arguments, single `Object` return. This uniform signature lets the evaluator call any built-in the same way it calls a user-defined `Function`.

| Method | Returns |
|---|---|
| `Type()` | `BUILTIN_OBJ` |
| `Inspect()` | `"builtin function"` (fixed string) |

---

### Array

Holds an ordered list of runtime objects. Elements are a plain `[]Object` slice so each element can be any type, including nested arrays.

```go
type Array struct {
    Elements []Object
}
```

| Method | Returns |
|---|---|
| `Type()` | `ARRAY_OBJ` |
| `Inspect()` | `"[el1, el2, el3]"` — calls `Inspect()` on each element, joined with `", "` |

`Array` does **not** implement `Hashable`.

---

### Hash

A key-value collection. The challenge is that Zen allows integers, booleans, and strings as keys — but `Object` is an interface and cannot be used directly as a Go map key. The solution is a two-level structure.

#### HashKey

A small comparable struct used as the actual Go map key:

```go
type HashKey struct {
    Type  ObjectType
    Value uint64
}
```

Both `Integer{1}` and `Boolean{true}` encode `uint64(1)` in their `Value` field, but the `Type` field (`"INTEGER"` vs `"BOOLEAN"`) makes them structurally distinct. Cross-type key collisions are impossible.

#### Hashable Interface

Any object type that can be used as a hash key must implement:

```go
type Hashable interface {
    HashKey() HashKey
}
```

| Type | Implements Hashable | Encoding |
|---|---|---|
| `Integer` | ✓ | Direct `uint64` cast of `Value` |
| `Boolean` | ✓ | `true → 1`, `false → 0` |
| `String` | ✓ | FNV-64a hash of content bytes |
| `Float` | ✗ | Floating point equality unreliable |
| `Array` | ✗ | Mutable |
| `Function` | ✗ | Not meaningful as a key |
| `Hash` | ✗ | Not meaningful as a key |

#### HashPair

Stores both the original key object and its value together:

```go
type HashPair struct {
    Key   Object
    Value Object
}
```

Without `HashPair`, `Inspect()` would only have a raw `uint64` to show — not the original `"name"` or `true`.

#### The Hash Object

```go
type Hash struct {
    Pairs map[HashKey]HashPair
}
```

| Method | Returns |
|---|---|
| `Type()` | `HASH_OBJ` |
| `Inspect()` | `"{key1: val1, key2: val2}"` — calls `Inspect()` on each original key and value |

Note: Go map iteration order is not guaranteed, so the output order of pairs in `Inspect()` is non-deterministic.

---

## `environment.go`

### What the Environment Is

The `Environment` is a runtime symbol table — it maps variable names to their current `Object` values. Every scope in a Zen program has its own `Environment`.

```go
type Environment struct {
    store map[string]Object
    outer *Environment
}
```

**`store`** — the map holding this scope's own bindings.

**`outer`** — pointer to the enclosing environment. This is the entire mechanism for lexical scoping. `nil` means this is the top-level global scope.

### Constructors

**`NewEnvironment()`** — creates a fresh environment with no outer scope. Called once at program startup to create the global environment.

**`NewEnclosedEnvironment(outer *Environment)`** — creates a new environment pointing to an existing one as its outer. Called every time a function is invoked.

### Methods

**`Get(name string) (Object, bool)`**

Searches `store` first. If not found and `outer` is not nil, delegates the lookup upward recursively. Returns `(Object, true)` on success, `(nil, false)` if the name is undefined at any level. This chain lookup is what allows a function body to read global variables without any special casing.

**`Set(name string, val Object) Object`**

Stores a binding in the **current** scope's `store`. Never walks up to outer scopes. Always writes locally. Returns the stored value — callers can write `return env.Set(name, val)` directly.

The decision to always write locally is intentional: `let x = 5` inside a function creates a local `x` and does not silently modify any outer `x` of the same name.

### Scoping In Practice

When a function call is evaluated:

```
Global Env:        { x: 5, add: Function }
                          ↑ outer
Function Call Env: { a: 3, b: 7 }
```

Inside the function body, looking up `a` finds it in the call env immediately. Looking up `x` misses the call env, walks up to global, finds it there. An undefined name misses both levels and becomes a runtime `Error`. When the function returns, the call env is discarded and garbage collected.

### How Closures Work

```
Global Env:           { makeCounter: Function }
                              ↑ outer
makeCounter Call Env: { count: 0 }
                              ↑ outer  ← captured in the closure
Returned Fn Env:      {}
```

The returned inner function holds a reference to `makeCounter`'s call env via its `Function.Env` field. Even after `makeCounter` returns and the evaluator discards that call frame, the env stays alive in memory because the closure holds the only remaining reference. Each subsequent call to the counter reads and modifies `count` in that captured env.

---

## Tests — `object_test.go`

All tests are in `package object` (white-box), giving direct access to struct fields and the package-level singletons `TRUE`, `FALSE`, and `NULL`.

The file uses a single local helper:

**`strContains(s, sub string) bool`** — manual substring search used in tests where exact output order is non-deterministic (Hash `Inspect()`, Function `Inspect()`). Implemented without importing `strings` to keep the helper minimal.

### Test Index

**Integer**

| Test | What it verifies |
|---|---|
| `TestIntegerType` | `Type()` returns `INTEGER_OBJ` |
| `TestIntegerInspect` | `0`, `42`, `-7`, `1000000` all format correctly via `%d` |
| `TestIntegerHashKey` | Same value → equal key; different value → unequal key |
| `TestIntegerHashKeyIncludesType` | `Integer{1}` and `Boolean{true}` produce different `HashKey`s despite both encoding `uint64(1)` |

**Float**

| Test | What it verifies |
|---|---|
| `TestFloatType` | `Type()` returns `FLOAT_OBJ` |
| `TestFloatInspect` | `3.14` → `"3.14"`, `0.0` → `"0"`, `-1.5` → `"-1.5"`, `100.0` → `"100"` — `%g` trailing-zero stripping |

**Boolean**

| Test | What it verifies |
|---|---|
| `TestBooleanSingletons` | `TRUE.Value=true`, `FALSE.Value=false`, they are distinct pointers |
| `TestBooleanType` | Both `TRUE` and `FALSE` return `BOOLEAN_OBJ` |
| `TestBooleanInspect` | `TRUE` → `"true"`, `FALSE` → `"false"` |
| `TestBooleanHashKey` | Each singleton's key is stable; `TRUE.HashKey() != FALSE.HashKey()` |

**Null**

| Test | What it verifies |
|---|---|
| `TestNullSingleton` | `NULL` is not Go nil; `Type()` = `NULL_OBJ`; `Inspect()` = `"null"` |

**String**

| Test | What it verifies |
|---|---|
| `TestStringType` | `Type()` returns `STRING_OBJ` |
| `TestStringInspect` | Returns `Value` directly without added quotes |
| `TestStringHashKey` | Same content → equal key; different content → unequal key |
| `TestStringHashKeyIncludesType` | Manufactures a deliberate collision: `Integer` whose value equals `String`'s hash output; confirms keys still differ via `Type` field |

**ReturnValue**

| Test | What it verifies |
|---|---|
| `TestReturnValueType` | `Type()` returns `RETURN_VALUE_OBJ` |
| `TestReturnValueInspect` | Delegates to wrapped `Integer{42}.Inspect()` → `"42"` |
| `TestReturnValueWrapsNull` | Correctly wraps `NULL` singleton → `"null"` |

**Error**

| Test | What it verifies |
|---|---|
| `TestErrorType` | `Type()` returns `ERROR_OBJ` |
| `TestErrorInspect` | `"ERROR: identifier not found: x"` format |
| `TestErrorInspectEmptyMessage` | Empty message produces `"ERROR: "` — prefix always present |

**Builtin**

| Test | What it verifies |
|---|---|
| `TestBuiltinType` | `Type()` returns `BUILTIN_OBJ` |
| `TestBuiltinInspect` | Returns `"builtin function"` |
| `TestBuiltinFnIsCallable` | `Fn` field actually executes when called |
| `TestBuiltinFnReceivesArgs` | Both arguments are passed through correctly in order |

**Array**

| Test | What it verifies |
|---|---|
| `TestArrayType` | `Type()` returns `ARRAY_OBJ` |
| `TestArrayInspectEmpty` | `[]Object{}` → `"[]"` |
| `TestArrayInspectThreeElements` | Three integers → `"[1, 2, 3]"` |
| `TestArrayInspectMixedTypes` | `Integer`, `Boolean`, `String` elements → `"[1, true, hello]"` |

**Hash**

| Test | What it verifies |
|---|---|
| `TestHashType` | `Type()` returns `HASH_OBJ` |
| `TestHashInspectEmpty` | Empty pairs map → `"{}"` |
| `TestHashInspectOneEntry` | Output contains original key string, value string, and `:` separator — order not asserted |
| `TestHashKeyEqualityAcrossTypes` | `Integer{1}`, `Boolean{true}`, `String{"1"}` all produce distinct `HashKey`s |

**Function**

| Test | What it verifies |
|---|---|
| `TestFunctionType` | `Type()` returns `FUNCTION_OBJ` |
| `TestFunctionInspectNoParams` | Output contains `"fn"` and `"()"` |
| `TestFunctionInspectWithParams` | Parameter names `"x"` and `"y"` appear in output |

**Environment — basic operations**

| Test | What it verifies |
|---|---|
| `TestEnvironmentSetAndGet` | `Set` then `Get` returns the correct value with `ok=true`; type-asserts to `*Integer` |
| `TestEnvironmentGetMissing` | Undefined name returns `ok=false` |
| `TestEnvironmentSetReturnsValue` | `Set` returns the same pointer that was stored |
| `TestEnvironmentOverwrite` | Re-setting the same name replaces the stored value |

**Environment — lexical scoping**

| Test | What it verifies |
|---|---|
| `TestEnclosedEnvironmentSeesOuterScope` | Inner env resolves a name only defined in outer |
| `TestEnclosedEnvironmentDoesNotLeakToOuter` | Variable set in inner is invisible to outer — scopes are one-way |
| `TestEnclosedEnvironmentShadowsOuterVariable` | Inner `x` shadows outer `x`; outer `x` is unchanged |
| `TestThreeLevelScopeChain` | Global → function → block: deepest scope sees all three levels; global cannot see inner levels |
| `TestClosureEnvironmentSurvivesOuterReturn` | Captured env stays alive after the direct reference is set to nil — closure holds it via `closureEnv` |

### Running the Tests

```bash
# All object tests
go test ./pkg/object/...

# Verbose
go test -v ./pkg/object/...

# Single test
go test -v -run TestClosureEnvironmentSurvivesOuterReturn ./pkg/object/...

# Full project
go test ./...
```