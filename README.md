<div align="center">

# Zen

**A scripting language built from scratch in Go.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Phase](https://img.shields.io/badge/Phase%201-Complete-brightgreen?style=flat)]()
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat)]()

</div>

---

Zen is a dynamically typed scripting language implemented entirely in Go as a learning project. The goal is to build a complete language from first principles — lexer, parser, AST, evaluator, and eventually a bytecode VM — understanding every piece of machinery along the way.

**Phase 1 (complete):** Tree-walking interpreter with an interactive REPL and file execution.  
**Phase 2 (planned):** Bytecode compiler and virtual machine for 5–10x faster execution.

```zen
let makeCounter = fn() {
    let count = 0
    fn() { count = count + 1; count }
}

let counter = makeCounter()
counter()   // 1
counter()   // 2
counter()   // 3
```

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Language Reference](#language-reference)
- [Built-in Functions](#built-in-functions)
- [Project Architecture](#project-architecture)
- [Running Tests](#running-tests)
- [Roadmap](#roadmap)

---

## Features

- **Types** — integers, floats, booleans, strings, null, arrays, hash maps
- **Variables** — `let` bindings with lexical scope
- **Operators** — arithmetic, comparison, logical, compound assignment (`+=`, `-=`, `*=`, `/=`), increment/decrement (`++`, `--`)
- **Control flow** — `if` / `else` expressions, `while` loops
- **Functions** — first-class, closures, recursion
- **Data structures** — arrays with index access, hash maps with string/integer/boolean keys
- **Built-ins** — `len`, `print`, `first`, `last`, `push`, `type`
- **Error handling** — runtime errors propagate cleanly without crashing the process
- **Interactive REPL** — persistent session state across lines
- **File execution** — run `.zen` scripts from the command line

---

## Installation

**Requirements:** Go 1.21 or later.

```bash
git clone https://github.com/Vamushi-gande/zenlang
cd zenlang
go build -o zen ./cmd/zen/main.go
```

To install to your `$GOPATH/bin` so `zen` is available anywhere:

```bash
go install ./cmd/zen/main.go
```

---

## Usage

### Interactive REPL

```bash
./zen
```

```
Welcome to Zen v0.1
Type 'exit' or 'quit' to leave

zen> let x = 10
zen> let double = fn(n) { n * 2 }
zen> double(x)
20
zen> exit
Goodbye!
```

Variables and functions defined in one line are available in all subsequent lines — the session shares a single persistent environment.

### Running a File

```bash
./zen script.zen
```

Parse errors and runtime errors are written to `stderr` with exit code 1, making Zen scripts composable in shell pipelines:

```bash
./zen script.zen && echo "success"
```

---

## Language Reference

### Variables

Variables are declared with `let`. There is no mutation through `let` — to update an existing variable use bare assignment.

```zen
let x = 10
let name = "zen"
let active = true
let nothing = null

x = 20          // update existing binding
x += 5          // compound assignment → 25
```

### Types

| Type    | Example                    |
|---------|----------------------------|
| Integer | `42`, `-7`, `0`            |
| Float   | `3.14`, `-0.5`             |
| Boolean | `true`, `false`            |
| String  | `"hello"`, `""`            |
| Null    | `null`                     |
| Array   | `[1, 2, 3]`                |
| Hash    | `{"key": "value"}`         |

Integer and float arithmetic is supported between compatible types. When one operand is a float the other is promoted automatically:

```zen
10 + 3      // 13    (integer)
10 + 3.0    // 13.0  (float)
10 / 3      // 3     (integer division)
```

### Operators

```zen
// Arithmetic
5 + 3    // 8
5 - 3    // 2
5 * 3    // 15
5 / 2    // 2  (integer division)
5 / 2.0  // 2.5

// Comparison
5 == 5   // true
5 != 3   // true
5 > 3    // true
5 < 3    // false
5 >= 5   // true
5 <= 4   // false

// Logical
true && false  // false
true || false  // true
!true          // false

// Compound assignment
let x = 10
x += 5   // 15
x -= 3   // 12
x *= 2   // 24
x /= 4   // 6

// Increment / decrement
++x      // 7
--x      // 6
```

### Strings

Strings support concatenation with `+` and equality comparison with `==` / `!=`.

```zen
let greeting = "Hello" + ", " + "world"
len(greeting)          // 13
greeting == "Hello"    // false
```

### If / Else

`if` is an expression — it produces a value. The `else` branch is optional; without it a falsy condition returns `null`.

```zen
let result = if (x > 10) { "big" } else { "small" }

// Only NULL and false are falsy. 0 and "" are truthy.
if (0) { print("this runs") }
if (null) { print("this does not") }
```

### While Loops

```zen
let i = 0
let sum = 0
while (i < 10) {
    sum = sum + i
    i += 1
}
sum   // 45
```

A `return` statement inside a while loop exits the enclosing function:

```zen
let findFirst = fn(arr, target) {
    let i = 0
    while (i < len(arr)) {
        if (arr[i] == target) { return i }
        i += 1
    }
    return -1
}
```

### Functions

Functions are first-class values. They are defined with `fn` and can be assigned to variables, passed as arguments, and returned from other functions.

```zen
let add = fn(a, b) { a + b }
add(3, 4)   // 7

// Immediately invoked
fn(x) { x * x }(5)   // 25
```

The last expression in a function body is its implicit return value. An explicit `return` exits the function early.

```zen
let abs = fn(n) {
    if (n < 0) { return -n }
    n
}
```

### Closures

Functions capture the environment in which they are defined. Inner functions can read and mutate variables from enclosing scopes.

```zen
let makeAdder = fn(x) {
    fn(y) { x + y }
}

let addTen = makeAdder(10)
addTen(5)    // 15
addTen(20)   // 30
```

Mutation of captured variables works through bare assignment (`=`), which walks the scope chain to find and update the original binding:

```zen
let makeCounter = fn() {
    let count = 0
    fn() { count = count + 1; count }
}

let c = makeCounter()
c()   // 1
c()   // 2
c()   // 3
```

### Recursion

```zen
let factorial = fn(n) {
    if (n <= 1) { return 1 }
    return n * factorial(n - 1)
}
factorial(10)   // 3628800

let fib = fn(n) {
    if (n <= 0) { return 0 }
    if (n == 1) { return 1 }
    return fib(n - 1) + fib(n - 2)
}
fib(10)   // 55
```

### Arrays

Arrays are ordered collections of any type. Index access uses `arr[i]`. Out-of-bounds indices return `null` rather than an error.

```zen
let arr = [1, "two", true, null]
arr[0]     // 1
arr[2]     // true
arr[99]    // null

let nums = [3, 1, 4, 1, 5]
first(nums)           // 3
last(nums)            // 5
len(nums)             // 5
push(nums, 9)         // [3, 1, 4, 1, 5, 9]  — original unchanged
```

Arrays are immutable through the built-in interface. `push` always returns a new array.

### Hash Maps

Hash maps store key-value pairs. Keys must be integers, booleans, or strings.

```zen
let person = {
    "name": "Alice",
    "age":  30,
    "active": true
}

person["name"]     // "Alice"
person["age"]      // 30
person["missing"]  // null

// Integer and boolean keys work too
let h = {1: "one", 2: "two", true: "yes"}
h[1]      // "one"
h[true]   // "yes"
```

---

## Built-in Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `len` | `len(str\|arr)` | Length of a string or array |
| `print` | `print(args...)` | Print arguments space-separated with a newline; returns `null` |
| `first` | `first(arr)` | First element of an array, or `null` if empty |
| `last` | `last(arr)` | Last element of an array, or `null` if empty |
| `push` | `push(arr, val)` | New array with `val` appended; original is unchanged |
| `type` | `type(val)` | String name of the value's type, e.g. `"INTEGER"`, `"STRING"` |

```zen
len("hello")           // 5
len([1, 2, 3])         // 3
print("x =", 42)       // prints: x = 42
first([10, 20, 30])    // 10
last([10, 20, 30])     // 30
push([1, 2], 3)        // [1, 2, 3]
type(3.14)             // "FLOAT"
type(null)             // "NULL"
```

---

## Project Architecture

Zen is built as a chain of packages, each depending only on those before it:

```
token → lexer → ast → parser → object → evaluator → repl → main
```

```
zen-lang/
├── cmd/
│   └── zen/
│       └── main.go          Entry point — REPL or file execution
└── pkg/
    ├── token/               Token type definitions
    ├── lexer/               Breaks source text into tokens
    ├── ast/                 AST node interfaces and types
    ├── parser/              Pratt parser — tokens → AST
    ├── object/              Runtime value types and environment
    ├── evaluator/           Tree-walking interpreter
    └── repl/                Interactive shell
```

### How a line of Zen executes

```
Source string
    │
    ▼ pkg/lexer
Tokens  [LET] [IDENT "x"] [ASSIGN] [INT "5"]  ...
    │
    ▼ pkg/parser  (Pratt parsing for operator precedence)
*ast.Program  →  [ LetStatement{ Name: "x", Value: IntegerLiteral{5} } ]
    │
    ▼ pkg/evaluator  (recursive tree walk)
object.Object  →  env.Set("x", Integer{5})  →  nil
```

### Package responsibilities

| Package | Responsibility |
|---------|---------------|
| `token` | Token type constants and the `Token` struct |
| `lexer` | Reads source character by character; produces a stream of `Token` values |
| `ast` | Defines the `Node`, `Statement`, and `Expression` interfaces; all AST node types |
| `parser` | Consumes the token stream; builds an `*ast.Program` using recursive descent and Pratt parsing for operator precedence |
| `object` | Defines the `Object` interface and all runtime types — `Integer`, `Float`, `String`, `Boolean`, `Null`, `Array`, `Hash`, `Function`, `Builtin`, `ReturnValue`, `Error` — plus the `Environment` (scope chain) |
| `evaluator` | Walks the AST recursively via `Eval(node, env)`; implements all operators, control flow, function calls, closures, and built-ins |
| `repl` | Wraps the pipeline in a Read-Eval-Print loop; uses `io.Reader`/`io.Writer` for testability |

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/evaluator/...

# Verbose output
go test -v ./...

# Run a single test
go test -v -run TestMakeCounter ./pkg/evaluator/...
```

All packages have test coverage. The evaluator alone has 60+ tests covering every operator, control flow path, closure behaviour, and built-in function.

---

## Roadmap

### Phase 1 — Tree-Walking Interpreter ✅
- Lexer, parser, AST
- Full expression evaluator
- Closures and recursion
- Arrays and hash maps
- Interactive REPL and file execution

### Phase 2 — Bytecode VM (planned)
- Opcode definitions (`pkg/code`)
- Compiler: AST → bytecode (`pkg/compiler`)
- Stack-based virtual machine (`pkg/vm`)
- Symbol table for variable resolution
- Target: 5–10x faster than the tree-walker
- Reuses `token`, `lexer`, `ast`, `parser` unchanged

---

## Learning Resources

This project follows the structure of:

- *Writing an Interpreter in Go* — Thorsten Ball
- *Writing a Compiler in Go* — Thorsten Ball
- *Crafting Interpreters* — Bob Nystrom (free at [craftinginterpreters.com](https://craftinginterpreters.com))