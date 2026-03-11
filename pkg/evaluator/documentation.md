# `pkg/evaluator` — Package Documentation

**Package:** `github.com/Vamshi-gande/zenlang/pkg/evaluator`
**Location:** `pkg/evaluator/`
**Files:** `evaluator.go`, `builtins.go`, `evaluator_test.go`

---

## Overview

The evaluator is the **execution engine** of Zen's Phase 1 interpreter. It implements a
**tree-walking interpreter**: it takes the `*ast.Program` produced by the parser, recursively
walks every node in the AST, and produces `object.Object` values as output. There is no
compilation step and no intermediate bytecode — execution is direct traversal.

```
Source Code
    │
    ▼
  Lexer  ──→  Tokens
    │
    ▼
 Parser  ──→  *ast.Program
    │
    ▼
Evaluator  ──→  object.Object   (uses object.Environment for variable scope)
```

Every piece of computation — arithmetic, variable lookup, function calls, closures, control
flow, data structures — flows through a single recursive entry point: `Eval`.

---

## File Structure

```
pkg/evaluator/
├── evaluator.go       → Eval dispatcher + all evaluation functions
├── builtins.go        → Built-in function map (len, print, first, last, push, type)
└── evaluator_test.go  → 70+ test cases covering all language features
```

---

## Dependencies

```
evaluator.go
    ├── pkg/ast     AST node types to type-switch on
    ├── pkg/object  Runtime value types + Environment
    └── fmt         Error message formatting

builtins.go
    ├── pkg/object
    └── fmt
```

---

## Architecture & Design Decisions

### Why Tree-Walking?

A tree-walking interpreter is the simplest possible execution model: no compiler, no VM, no
register allocation. `Eval` is called on a node, evaluates its children by calling `Eval`
recursively, and returns an `object.Object`. This makes the code easy to understand and
extend — adding a new language feature means writing one new case in the type switch and one
new `evalXxx` function.

The trade-off is performance. Every expression re-traverses the AST on every execution, which
is 5–10x slower than a bytecode VM. For Phase 1 this is fine; Phase 2 (the bytecode compiler
and VM) replaces this package for production use.

### The Object System as the Return Type

Every evaluation function returns `object.Object`, a uniform interface. This means:

- Errors are just another kind of object (`*object.Error`). They do not use Go's `error`
  interface or panic — they flow through the same pipeline as normal values.
- `return` is implemented by wrapping a value in `*object.ReturnValue` and letting it bubble
  up through the call stack naturally, without modifying any calling function's control flow.
- `null` is a legitimate first-class value, not a nil pointer.

### Singletons

Three package-level singletons are declared in `pkg/object` and used throughout:

| Singleton      | Type                    | Why a singleton?                               |
|----------------|-------------------------|------------------------------------------------|
| `object.TRUE`  | `*object.Boolean{true}` | Boolean results never need allocation          |
| `object.FALSE` | `*object.Boolean{false}`| Boolean results never need allocation          |
| `object.NULL`  | `*object.Null`          | null has no value — one instance is sufficient |

Because `TRUE` and `FALSE` are singletons, `==` and `!=` on booleans work via **pointer
comparison** (`left == right`). This is both fast and correct — you never end up with two
different `*object.Boolean{true}` instances that compare unequal.

### `env.Set` vs `env.Update` — the Scope Rule

This distinction is the most important design decision in the evaluator.

`env.Set(name, val)` always writes to the **current** (innermost) scope. It is used by:
- `let x = expr` — creates a new binding in the current scope
- function parameter binding — each call gets its own local binding

`env.Update(name, val)` walks the scope chain and modifies the binding in the **first scope
that already holds the name**. It is used by bare assignment `x = expr`.

This split is what makes closures with mutation work:

```zen
let makeCounter = fn() {
    let count = 0             -- env.Set in makeCounter's scope
    fn() {
        count = count + 1     -- env.Update: walks up, finds count in makeCounter's scope, mutates it
        count
    }
}
```

If `=` used `env.Set` instead of `env.Update`, each call to the inner function would create
a brand-new `count` local to the closure's own scope, resetting to 0 every time.

### `evalProgram` vs `evalBlockStatement` — the ReturnValue Rule

Both functions evaluate a list of statements in sequence. Their difference is how they handle
`*object.ReturnValue`:

- **`evalProgram`** unwraps the `ReturnValue` and returns the inner value. It is the
  top-level execution boundary.
- **`evalBlockStatement`** passes `ReturnValue` **upward without unwrapping**. It is used
  for function bodies, `if`/`else` branches, and `while` bodies.

Unwrapping only ever happens at the **function call boundary** (`unwrapReturnValue` inside
`applyFunction`). This means `return` inside any level of nesting inside a function correctly
exits the whole function:

```
fn(n) {
    if (n > 0) {            -- evalBlockStatement: sees ReturnValue, passes it up
        return n * 2        -- ReturnValue{n*2} created here
    }
    return 0
}                           -- applyFunction calls unwrapReturnValue here
```

If `evalBlockStatement` unwrapped `ReturnValue`, the `if` body would consume it and
execution would fall through to `return 0`.

---

## Entry Point — `Eval`

```go
func Eval(node ast.Node, env *object.Environment) object.Object
```

The single recursive dispatch function. It type-switches on the incoming AST node and routes
to the appropriate handler. Nodes that have children call `Eval` recursively on those children.

Returns `nil` only for `*ast.LetStatement` (which has no expression value). Every other node
returns a non-nil `object.Object`.

### Complete Dispatch Table

| AST Node                   | Handler / Direct return                 | Notes                                          |
|----------------------------|-----------------------------------------|------------------------------------------------|
| `*ast.Program`             | `evalProgram`                           | Unwraps `ReturnValue` at the top level         |
| `*ast.ExpressionStatement` | `Eval(node.Expression, env)`            | Transparent wrapper, no side effects           |
| `*ast.BlockStatement`      | `evalBlockStatement`                    | Passes `ReturnValue` upward without unwrapping |
| `*ast.ReturnStatement`     | `evalReturnStatement`                   | Wraps result in `*object.ReturnValue`          |
| `*ast.LetStatement`        | `evalLetStatement`                      | Binds name in env; returns nil                 |
| `*ast.WhileStatement`      | `evalWhileStatement`                    | Loop; re-evaluates condition each iteration    |
| `*ast.IntegerLiteral`      | `&object.Integer{Value: node.Value}`    | Direct allocation                              |
| `*ast.FloatLiteral`        | `&object.Float{Value: node.Value}`      | Direct allocation                              |
| `*ast.StringLiteral`       | `&object.String{Value: node.Value}`     | Direct allocation                              |
| `*ast.BooleanLiteral`      | `nativeBoolToBooleanObject(node.Value)` | Returns `TRUE`/`FALSE` singleton               |
| `*ast.NullLiteral`         | `object.NULL`                           | Returns singleton directly                     |
| `*ast.Identifier`          | `evalIdentifier`                        | env chain → builtins → error                   |
| `*ast.PrefixExpression`    | `evalPrefixExpression`                  | `!`, `-`, `++`, `--`                           |
| `*ast.InfixExpression`     | `evalInfixExpression`                   | All binary operators including `=`             |
| `*ast.IfExpression`        | `evalIfExpression`                      | Evaluates one branch or returns NULL           |
| `*ast.FunctionLiteral`     | `&object.Function{Parameters, Body, Env: env}` | Captures current env for closure       |
| `*ast.CallExpression`      | `evalCallExpression`                    | Evaluates callee + args, applies function      |
| `*ast.ArrayLiteral`        | `evalArrayLiteral`                      | Evaluates all elements left-to-right           |
| `*ast.IndexExpression`     | `evalIndexExpression`                   | Dispatches to array or hash handler            |
| `*ast.HashLiteral`         | `evalHashLiteral`                       | Evaluates all key-value pairs                  |

---

## Statement Evaluation

### `evalProgram`

```go
func evalProgram(program *ast.Program, env *object.Environment) object.Object
```

Iterates `program.Statements`. Returns the value of the last statement evaluated. Stops early
if it encounters a `*object.ReturnValue` (unwraps and returns the inner value) or a
`*object.Error` (returns the error immediately).

---

### `evalBlockStatement`

```go
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object
```

Identical to `evalProgram` except `*object.ReturnValue` is **not** unwrapped — it is
returned as-is so it can bubble up to the function call boundary. Errors still stop
execution immediately.

---

### `evalLetStatement`

```go
func evalLetStatement(node *ast.LetStatement, env *object.Environment) object.Object
```

1. Evaluates `node.Value` (the right-hand expression).
2. Propagates any error.
3. Calls `env.Set(node.Name.Value, val)` — binds the name in the **current** scope.
4. Returns `nil` — let statements have no expression value.

---

### `evalReturnStatement`

```go
func evalReturnStatement(node *ast.ReturnStatement, env *object.Environment) object.Object
```

Evaluates `node.ReturnValue`, propagates errors, wraps the result in
`&object.ReturnValue{Value: val}`. The wrapper travels up through `evalBlockStatement` calls
until `unwrapReturnValue` strips it at the function call boundary.

---

### `evalWhileStatement`

```go
func evalWhileStatement(node *ast.WhileStatement, env *object.Environment) object.Object
```

```
loop:
  evaluate condition
  if not truthy → break
  evaluate body block
  if body returned ReturnValue or Error → propagate, stop loop
repeat
return NULL
```

Key points:
- The condition is fully re-evaluated from scratch on every iteration.
- The loop does **not** consume `ReturnValue` — it propagates it so the enclosing function
  exits correctly.
- Returns `NULL` when the loop exits normally (condition became falsy).

---

## Identifier Lookup — `evalIdentifier`

```go
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object
```

Resolution order:

1. `env.Get(node.Value)` — walks the scope chain from innermost to outermost
2. `builtins[node.Value]` — the package-level map in `builtins.go`
3. `newError("identifier not found: %s", node.Value)`

User-defined variables always shadow builtins. It is valid (though inadvisable) to write
`let len = 5` and shadow the built-in `len`.

---

## Prefix Expressions — `evalPrefixExpression`

```go
func evalPrefixExpression(node *ast.PrefixExpression, env *object.Environment) object.Object
```

Evaluates the right operand first. If it produces an error, that error propagates immediately
without applying the operator.

| Operator | Handler                   | Behaviour                                                                 | Error cases                         |
|----------|---------------------------|---------------------------------------------------------------------------|-------------------------------------|
| `!`      | `evalBangOperator`        | `FALSE`→`TRUE`, `TRUE`→`FALSE`, `NULL`→`TRUE`, anything else→`FALSE`     | none                                |
| `-`      | `evalMinusPrefixOperator` | Negates integer value                                                     | non-integer type                    |
| `++`     | `evalIncrementPrefix`     | Right must be `*ast.Identifier`; increments by 1; updates env; returns new value | non-identifier; non-integer  |
| `--`     | `evalDecrementPrefix`     | Right must be `*ast.Identifier`; decrements by 1; updates env; returns new value | non-identifier; non-integer  |

> **Note:** `++` and `--` are **prefix-only**. `--5` lexes as `DEC` applied to the literal
> `5`, which has no identifier to update — this is a runtime error. To double-negate, write
> `- -5` (two minus tokens separated by a space).

---

## Infix Expressions — `evalInfixExpression`

```go
func evalInfixExpression(node *ast.InfixExpression, env *object.Environment) object.Object
```

Both operands are evaluated before any type checking. The dispatch order is:

1. **`=` (bare assignment)** — checked first, before type dispatch
2. **Both `INTEGER`** → `evalIntegerInfixExpression`
3. **Either `FLOAT`** → `evalFloatInfixExpression` (promotes integer side to `float64`)
4. **Both `STRING`** → `evalStringInfixExpression`
5. **`==` / `!=`** → pointer comparison (correct for booleans and null via singletons)
6. **`&&` / `||`** → `isTruthy`-based evaluation
7. **Types differ** → `"type mismatch: <T1> <op> <T2>"` error
8. **Default** → `"unknown operator: <T> <op> <T>"` error

---

### `evalAssignment` — bare `x = expr`

```go
func evalAssignment(node *ast.InfixExpression, val object.Object, env *object.Environment) object.Object
```

The left operand must be an `*ast.Identifier`; anything else is a runtime error. Calls
`env.Update(name, val)` which walks the scope chain to find and mutate the existing binding
in place. Returns the assigned value.

This is the mechanism behind closure mutation — see the `env.Set` vs `env.Update` section above.

---

### `evalIntegerInfixExpression`

Handles all infix operations where both operands are integers.

| Operators                       | Behaviour                                                          |
|---------------------------------|--------------------------------------------------------------------|
| `+` `-` `*`                     | Standard arithmetic; returns new `*object.Integer`                 |
| `/`                             | Integer division; `"division by zero"` error if right is 0         |
| `<` `>` `<=` `>=` `==` `!=`    | Comparison; returns `TRUE`/`FALSE` singleton                       |
| `+=` `-=` `*=` `/=`            | Compound assignment; left must be `*ast.Identifier`; calls `env.Set`; returns new value |

---

### `evalFloatInfixExpression`

Invoked when at least one operand is a `FLOAT`. Promotes the non-float operand to `float64`.
Supports `+`, `-`, `*`, `/`, `<`, `>`, `<=`, `>=`, `==`, `!=`. Does not support compound
assignments. Division by zero returns an error.

---

### `evalStringInfixExpression`

Supports only three operators:

| Operator | Behaviour                         |
|----------|-----------------------------------|
| `+`      | Concatenates both strings         |
| `==`     | Value comparison; returns boolean |
| `!=`     | Value comparison; returns boolean |

Any other operator returns `"unknown operator: STRING <op> STRING"`.

---

## Truthiness — `isTruthy`

```go
func isTruthy(obj object.Object) bool
```

Zen uses a **simple, explicit** truthiness model. Only two values are falsy:

| Value               | Truthy?                                  |
|---------------------|------------------------------------------|
| `NULL`              | **false**                                |
| `false`             | **false**                                |
| `true`              | true                                     |
| `0` (integer zero)  | **true** — unlike JavaScript / Python    |
| `""` (empty string) | **true**                                 |
| `[]` (empty array)  | **true**                                 |
| any other value     | true                                     |

The implementation is a pointer switch on the three singletons. It never inspects value
fields of integers, strings, or arrays — keeping truthiness fast, predictable, and
type-independent.

---

## Control Flow — `evalIfExpression`

```go
func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object
```

1. Evaluates `node.Condition`; propagates any error before touching either branch.
2. Calls `isTruthy` on the result.
3. If truthy → evaluates `node.Consequence`.
4. If falsy and `node.Alternative != nil` → evaluates `node.Alternative`.
5. If falsy and no alternative → returns `object.NULL`.

---

## Functions & Closures

### `evalCallExpression`

```go
func evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object
```

Three steps:

1. `Eval(node.Function, env)` — evaluates the callee to a `*object.Function` or
   `*object.Builtin`. Propagates any error.
2. `evalExpressions(node.Arguments, env)` — evaluates arguments **left-to-right**. Stops on
   the first error; returns a single-element slice `[error]`.
3. `applyFunction(fn, args)` — calls the function.

---

### `evalExpressions`

```go
func evalExpressions(exprs []ast.Expression, env *object.Environment) []object.Object
```

Evaluates each expression in order. If any produces an error, returns
`[]object.Object{error}` immediately. The caller detects this with
`len(args) == 1 && isError(args[0])`.

---

### `applyFunction`

```go
func applyFunction(fn object.Object, args []object.Object) object.Object
```

Type-switches on `fn`:

- `*object.Function` → `extendFunctionEnv(fn, args)` + `Eval(fn.Body, extendedEnv)` + `unwrapReturnValue`
- `*object.Builtin` → calls `fn.Fn(args...)` directly
- anything else → `"not a function: <type>"` error

---

### `extendFunctionEnv`

```go
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment
```

Creates a new enclosed environment with `fn.Env` (the scope captured at **definition** time)
as the outer. Then binds each parameter name to its argument value with `env.Set`.

**Why `fn.Env` and not the caller's env?**
Using the defining environment rather than the calling environment gives Zen **lexical
scoping**. A function sees variables from the scope where it was *defined*, not where it is
*called*. This is also what makes closures work — the inner function retains a live reference
to the outer function's environment even after the outer call has returned.

---

### `unwrapReturnValue`

```go
func unwrapReturnValue(obj object.Object) object.Object
```

If `obj` is a `*object.ReturnValue`, returns its inner value. Otherwise passes `obj` through
unchanged. Called only at the function call boundary inside `applyFunction`.

---

## Data Structures

### `evalArrayLiteral`

Calls `evalExpressions` on `node.Elements`, wraps the result in `&object.Array{Elements: ...}`.
Any element error stops evaluation and propagates.

---

### `evalIndexExpression`

Evaluates `node.Left` (the collection) and `node.Index` (the key), then dispatches:

| Collection type | Index type | Handler                     |
|-----------------|------------|-----------------------------|
| `ARRAY`         | `INTEGER`  | `evalArrayIndexExpression`  |
| `HASH`          | any        | `evalHashIndexExpression`   |
| anything else   | any        | `"index operator not supported: <type>"` error |

---

### `evalArrayIndexExpression`

```go
func evalArrayIndexExpression(array, index object.Object) object.Object
```

Returns `arr.Elements[idx]`. Out-of-bounds indices (negative or beyond `len-1`) return
`object.NULL` rather than an error — a deliberate design choice for friendlier scripting.

---

### `evalHashIndexExpression`

```go
func evalHashIndexExpression(hash, index object.Object) object.Object
```

The index must implement `object.Hashable`. Supported hashable types:

| Type      | Hash method                           |
|-----------|---------------------------------------|
| `INTEGER` | Direct `uint64` cast of `int64` value |
| `BOOLEAN` | `true` → 1, `false` → 0              |
| `STRING`  | FNV-64a hash of the string bytes      |

Non-hashable types (float, array, function, etc.) produce `"unusable as hash key: <type>"`.
A valid but absent key returns `object.NULL`.

> **Design note:** `HashKey` stores both `Type ObjectType` and `Value uint64`. The `Type`
> field prevents cross-type collisions — `Integer{1}` and `Boolean{true}` both encode to
> `uint64(1)` but have different `HashKey` structs, so they never collide.

---

### `evalHashLiteral`

Iterates `node.Pairs` (a `map[ast.Expression]ast.Expression`). For each pair:

1. Evaluates the key expression; propagates any error.
2. Asserts the key implements `object.Hashable`; returns an error if not.
3. Evaluates the value expression; propagates any error.
4. Stores `object.HashPair{Key: key, Value: value}` in the result map keyed by `HashKey`.

---

## Error Handling

### Propagation Model

The evaluator uses a **propagation model** rather than Go panics or exceptions.
`*object.Error` is just another `object.Object`. After almost every `Eval` call, the result
is checked with `isError`. If true, the error is returned immediately:

```go
val := Eval(node.Right, env)
if isError(val) {
    return val   // stop here, propagate upward
}
```

Errors travel back up the call stack through normal return values, stopping at every level.
The error that reaches `evalProgram` is the final output.

---

### Helper Functions

**`newError(format string, a ...interface{}) *object.Error`**
Creates `&object.Error{Message: fmt.Sprintf(format, a...)}`. Used everywhere a runtime error
must be produced.

**`isError(obj object.Object) bool`**
Returns `true` if `obj` is non-nil and `obj.Type() == object.ERROR_OBJ`. The nil check is
necessary because `evalLetStatement` returns nil.

**`nativeBoolToBooleanObject(input bool) *object.Boolean`**
Returns `object.TRUE` or `object.FALSE` singleton. Never allocates. Every boolean-producing
operation calls this rather than `&object.Boolean{...}`.

---

### Runtime Error Reference

| Error message                                     | Cause                                                    |
|---------------------------------------------------|----------------------------------------------------------|
| `identifier not found: <n>`                       | Name not in env chain or builtins                        |
| `division by zero`                                | Integer or float division by 0                           |
| `unknown operator: -<TYPE>`                       | Minus prefix applied to non-integer                      |
| `unknown operator: <T> <op> <T>`                  | Operator not defined for that type                       |
| `type mismatch: <T1> <op> <T2>`                   | Infix with incompatible types                            |
| `not a function: <type>`                          | Call expression on a non-function value                  |
| `unusable as hash key: <type>`                    | Non-hashable type used as a map key                      |
| `index operator not supported: <type>`            | Index applied to non-array, non-hash value               |
| `assignment target must be an identifier`         | Left side of `=` is not a variable name                  |
| `compound assignment requires an identifier`      | Left side of `+=` etc. is not a variable name            |
| `operator ++ requires an identifier`              | `++` applied to a literal                                |
| `operator -- requires an identifier`              | `--` applied to a literal                                |
| `wrong number of arguments to <builtin>`          | Built-in called with wrong arg count                     |
| `argument to <builtin> not supported, got <type>` | Built-in called with wrong argument type                 |

---

## `builtins.go` — Built-in Functions

The `builtins` variable is a package-level `map[string]*object.Builtin`. `evalIdentifier`
checks it after failing to find a name in the environment, so builtins are always in scope
and can be shadowed by user `let` bindings.

### Built-in Reference

| Name    | Signature          | Returns                                  | Errors                               |
|---------|--------------------|------------------------------------------|--------------------------------------|
| `len`   | `len(str\|arr)`    | `INTEGER` length of string or array      | Wrong arg count; unsupported type    |
| `print` | `print(args...)`   | `NULL`                                   | none                                 |
| `first` | `first(arr)`       | First element, or `NULL` if empty        | Wrong arg count; non-array           |
| `last`  | `last(arr)`        | Last element, or `NULL` if empty         | Wrong arg count; non-array           |
| `push`  | `push(arr, val)`   | New array with `val` appended            | Wrong arg count; non-array first arg |
| `type`  | `type(val)`        | `STRING` type name e.g. `"INTEGER"`      | Wrong arg count                      |

**`push` is non-mutating.** It allocates a new slice, copies original elements, appends
`val`, and returns a new `*object.Array`. The original array is unchanged.

**`print`** calls `Inspect()` on each argument, joins with spaces, and calls `fmt.Println`.
Returns `NULL`.

**`type`** returns the `ObjectType` constant directly: `"INTEGER"`, `"FLOAT"`, `"STRING"`,
`"BOOLEAN"`, `"NULL"`, `"ARRAY"`, `"HASH"`, `"FUNCTION"`, `"BUILTIN"`.

---

## Environment & Scope — `pkg/object.Environment`

The environment is a **linked list of scopes**:

```go
type Environment struct {
    store map[string]Object
    outer *Environment   // nil for global scope
}
```

### Methods

| Method                          | Behaviour                                                                               |
|---------------------------------|-----------------------------------------------------------------------------------------|
| `NewEnvironment()`              | Creates a global scope with `outer = nil`                                               |
| `NewEnclosedEnvironment(outer)` | Creates a new scope linked to `outer` — used for every function call                   |
| `Get(name)`                     | Walks outer chain until name found; returns `(value, true)` or `("", false)`           |
| `Set(name, val)`                | Writes to **current** scope only; used by `let` and parameter binding                  |
| `Update(name, val)`             | Walks outer chain and modifies the **first** scope holding `name`; used by bare `=`; creates in current scope if not found |

### Scope Lifetime

A new enclosed environment is created for every function **call** — not definition. Recursive
functions get fresh local scopes on each call and do not share state between frames.

The function's **definition** environment (`fn.Env`) is captured once when the
`*ast.FunctionLiteral` is evaluated. It persists as long as the `*object.Function` object
is alive, even after the defining function has returned — this is what gives closures their
memory.

---

## Test Coverage — `evaluator_test.go`

All tests use `testEval(input string) object.Object`, which runs the full pipeline:

```
NewLexer(input) → NewParser(lexer) → ParseProgram() → NewEnvironment() → Eval(program, env)
```

Typed assertion helpers verify the returned object:

| Helper                         | Asserts                                                    |
|--------------------------------|------------------------------------------------------------|
| `assertInteger(t, obj, want)`  | `obj` is `*object.Integer` with `Value == want`            |
| `assertFloat(t, obj, want)`    | `obj` is `*object.Float` with `Value` within 1e-9 of `want` |
| `assertBoolean(t, obj, want)`  | `obj` is `*object.Boolean` with `Value == want`            |
| `assertString(t, obj, want)`   | `obj` is `*object.String` with `Value == want`             |
| `assertNull(t, obj)`           | `obj` is the `object.NULL` singleton                       |
| `assertError(t, obj, substr)`  | `obj` is `*object.Error` whose `Message` contains `substr` |

### Full Test Index

| Test                                  | What it verifies                                                       |
|---------------------------------------|------------------------------------------------------------------------|
| `TestIntegerExpression`               | Arithmetic `+` `-` `*` `/` with precedence                            |
| `TestIntegerDivisionByZero`           | `5 / 0` produces `"division by zero"` error                           |
| `TestFloatExpression`                 | Float arithmetic and int-float promotion                               |
| `TestFloatComparisons`                | `<` `>` `<=` `>=` `==` `!=` on floats                                 |
| `TestBooleanExpression`               | Comparison operators returning boolean values                          |
| `TestLogicalOperators`                | `&&` and `\|\|` with all combinations                                 |
| `TestBangOperator`                    | `!true`, `!false`, `!null`, `!5`                                       |
| `TestMinusPrefix`                     | `-5`, `-10`, `- -5` (double negation), `-true` → error                 |
| `TestIfElseExpression`                | Truthy/falsy condition; missing else returns null                      |
| `TestIfElseWithExpressionCondition`   | Non-boolean (integer) condition                                        |
| `TestReturnStatements`                | `return` at various nesting levels                                     |
| `TestReturnInsideNestedBlocks`        | `return` inside `if` inside a function exits the whole function        |
| `TestErrorHandling`                   | Type mismatch and unknown operator error message format                |
| `TestErrorStopsExecution`             | An error prevents subsequent statements from running                   |
| `TestNullLiteral`                     | `null` evaluates to the `NULL` singleton                               |
| `TestLetStatements`                   | `let` binds correctly; subsequent use returns the bound value          |
| `TestUndefinedVariable`               | Undefined name produces `"identifier not found"` error                 |
| `TestCompoundAssignment`              | `+=` `-=` `*=` `/=` update variables and return new value              |
| `TestIncrementDecrement`              | `++x` and `--x` update variables and return new value                 |
| `TestStringLiteral`                   | String value is stored and returned correctly                          |
| `TestStringConcatenation`             | `+` on strings produces concatenated result                            |
| `TestStringEquality`                  | `==` and `!=` on strings compare by value                              |
| `TestStringOperatorError`             | `-` on strings produces a type error                                   |
| `TestFunctionLiteral`                 | `fn(x){x}` produces `*object.Function` with correct params and body   |
| `TestFunctionCall`                    | Simple call; argument binding; returns body value                      |
| `TestFunctionCallNoArgs`              | `fn() { 42 }()` evaluates to 42                                        |
| `TestClosure`                         | Inner function reads variable from outer scope                         |
| `TestMakeCounter`                     | Closure mutation: `count = count + 1` updates captured environment    |
| `TestClosureCapturesOuterVariable`    | Outer variable read (not mutated) inside returned closure              |
| `TestFactorial`                       | Recursive `factorial(5)` → 120                                         |
| `TestFactorialZero`                   | Base case `factorial(0)` → 1                                           |
| `TestFibonacci`                       | Recursive `fib(10)` → 55                                               |
| `TestWhileLoop`                       | Loop increments counter to expected value using `x = x + 1`           |
| `TestWhileLoopNeverExecutes`          | `while(false)` body never runs; original value preserved               |
| `TestWhileLoopReturnInsideBody`       | `return` inside while exits the enclosing function correctly           |
| `TestArrayLiteral`                    | `[1, 2, 3]` — all three elements evaluated correctly                   |
| `TestArrayIndexExpression`            | `arr[0]`, `arr[2]` return correct elements                             |
| `TestArrayIndexOutOfBounds`           | `arr[99]` and `arr[-1]` return null (no error)                         |
| `TestArrayEmptyLiteral`               | `[]` produces an `*object.Array` with zero elements                    |
| `TestHashLiteralStringKeys`           | `{"one": 1}["one"]` → 1                                               |
| `TestHashLiteralIntegerKeys`          | `{1: "one"}[1]` → `"one"`                                             |
| `TestHashLiteralBooleanKeys`          | `{true: 1, false: 0}[true]` → 1                                       |
| `TestHashMissingKey`                  | Missing key returns null                                               |
| `TestHashUnusableKeyError`            | Float key → `"unusable as hash key"` error                             |
| `TestBuiltinLen`                      | `len("hello")` → 5; `len([1,2,3])` → 3                                |
| `TestBuiltinLenErrors`                | Wrong arg count; unsupported type                                      |
| `TestBuiltinFirst`                    | First element returned; empty array → null                             |
| `TestBuiltinFirstErrors`              | Wrong arg count; non-array argument                                    |
| `TestBuiltinLast`                     | Last element returned; empty array → null                              |
| `TestBuiltinLastErrors`               | Wrong arg count; non-array argument                                    |
| `TestBuiltinPush`                     | New array with element appended; correct length and value              |
| `TestBuiltinPushDoesNotMutateOriginal`| Original array unchanged after `push`                                  |
| `TestBuiltinPushOntoEmpty`            | `push([], 1)` → `[1]`                                                 |
| `TestBuiltinPushErrors`               | Wrong arg count; non-array first argument                              |
| `TestBuiltinType`                     | `type(5)` → `"INTEGER"`, `type("x")` → `"STRING"`, etc.               |
| `TestBuiltinTypeErrors`               | Wrong arg count                                                        |
| `TestCallingNonFunction`              | `5()` → `"not a function: INTEGER"` error                              |
| `TestTruthiness`                      | Only NULL and FALSE are falsy; 0 and `""` are truthy                   |
| `TestSumArray`                        | Integration: sum array elements using while loop and indexing          |
| `TestMapFunction`                     | Integration: map higher-order function applies fn to each element      |
| `TestHigherOrderFunctions`            | `apply(fn, value)` pattern with function passed as argument            |
| `TestFunctionReturningFunction`       | Function that returns another function; outer arg visible inside       |

### Running Tests

```bash
# Run all evaluator tests
go test ./pkg/evaluator/...

# Verbose output (shows each test name)
go test -v ./pkg/evaluator/...

# Run a single test by name
go test -v -run TestMakeCounter ./pkg/evaluator/...

# Run the entire project
go test ./...
```