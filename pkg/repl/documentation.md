# `pkg/repl` — Package Documentation

**Package:** `github.com/Vamushi-gande/zenlang/pkg/repl`
**Location:** `pkg/repl/`
**Files:** `repl.go`, `repl_test.go`
**Related:** `cmd/zen/main.go`

---

## Overview

The REPL is the **user-facing interface** for Phase 1. It is the final package in the
dependency chain — the point where every other package converges into something a person can
actually interact with. REPL stands for **Read-Eval-Print Loop**, which describes exactly
what happens on every iteration:

```
User Input  →  Read  →  Eval  →  Print  →  Loop back
```

The REPL also handles **file execution mode**, which runs a `.zen` source file through the
same pipeline once without an interactive loop. This is coordinated by `cmd/zen/main.go`.

---

## File Structure

```
pkg/repl/
├── repl.go        → REPL struct, Start loop, eval pipeline
└── repl_test.go   → 22 tests using captured I/O

cmd/zen/
└── main.go        → Entry point; routes to REPL or file execution
```

---

## Dependencies

```
repl.go
    ├── pkg/lexer      tokenises each input line
    ├── pkg/parser     parses tokens into an AST
    ├── pkg/evaluator  walks the AST and produces a result
    ├── pkg/object     runtime value types + Environment
    ├── bufio          line-by-line scanning
    ├── fmt            printing prompt and results
    └── io             Reader/Writer interfaces for testability

main.go
    ├── pkg/repl
    ├── pkg/lexer
    ├── pkg/parser
    ├── pkg/evaluator
    ├── pkg/object
    └── os             file reading, stderr, exit codes
```

The full dependency chain for the project is:

```
token → lexer → ast → parser → object → evaluator → repl → main
```

---

## Architecture & Design Decisions

### Why `io.Reader` / `io.Writer` instead of `os.Stdin` / `os.Stdout`?

The most important design decision in this package is that the `REPL` struct holds
`io.Reader` and `io.Writer` interfaces rather than hardcoded standard streams. This has two
consequences:

**Testability.** Every test in `repl_test.go` creates a `strings.NewReader` for input and a
`bytes.Buffer` for output. The entire REPL — including the prompt, error output, and result
printing — is exercised without any terminal interaction. Tests are deterministic and fast.

**Flexibility.** `repl.New` can be wired to any reader/writer pair. A future embedding use
case could connect the REPL to a network socket, a file, or a pipe without any change to
`repl.go`.

### Persistent Environment

The `*object.Environment` is created **once** in `New()` and stored on the struct. It is
passed to `evaluator.Eval` on every iteration. This is what makes variables defined in one
input line visible in subsequent lines — the entire session shares one scope chain.

If a fresh environment were created inside `eval()` instead, every input line would be an
isolated program with no memory of anything typed before. The REPL would be useless for
interactive use.

### NULL and nil Suppression

The `eval()` method prints nothing when the result is `nil` or `object.NULL`. There are two
separate cases:

- `evalLetStatement` in the evaluator returns **Go nil** — let statements produce no value.
- Expressions that explicitly evaluate to `null` return **`object.NULL`** (the singleton).

Suppressing both keeps the session clean. Without this, every variable declaration would
print `null`, which would be confusing:

```
zen> let x = 10     ← should produce no output
null                ← bad — this would print without suppression
zen> x + 5
15
```

### Error Recovery — Never Exit on User Mistakes

Both parse errors and runtime errors are printed and then discarded. The loop always
continues to the next iteration. The REPL only exits on EOF or an explicit `exit`/`quit`
command. This matches the expected behaviour of any interactive shell — a typo should never
terminate the session.

---

## `repl.go` — Function Reference

### `PROMPT`

```go
const PROMPT = "zen> "
```

The string printed at the start of every iteration before reading input. Displayed
unconditionally — even after errors — so the user always knows the shell is waiting.

---

### `REPL` struct

```go
type REPL struct {
    in  io.Reader
    out io.Writer
    env *object.Environment
}
```

| Field | Type | Purpose |
|-------|------|---------|
| `in`  | `io.Reader` | Source of input lines; `os.Stdin` in production, `strings.Reader` in tests |
| `out` | `io.Writer` | Destination for all output; `os.Stdout` in production, `bytes.Buffer` in tests |
| `env` | `*object.Environment` | Shared scope across all loop iterations |

---

### `New`

```go
func New(in io.Reader, out io.Writer) *REPL
```

Creates a `REPL` with a fresh global environment. The environment is initialised here and
never replaced — it accumulates all bindings for the lifetime of the session.

---

### `Start`

```go
func (r *REPL) Start()
```

The main loop. Runs until one of three termination conditions:

| Condition | Behaviour |
|-----------|-----------|
| `scanner.Scan()` returns false | EOF or read error — return silently |
| Input line is `"exit"` or `"quit"` | Print `"Goodbye!"` and return |
| Any other condition | Call `r.eval(line)` and loop |

The loop always prints the prompt **before** reading, so the user sees `zen> ` even on the
first iteration.

---

### `eval`

```go
func (r *REPL) eval(input string)
```

Runs a single line of Zen source through the full pipeline:

```
input string
    │
    ▼
lexer.NewLexer(input)          → *lexer.Lexer
    │
    ▼
parser.NewParser(l)            → *parser.Parser
    │
    ▼
p.ParseProgram()               → *ast.Program
    │
    ├── if p.Errors() > 0 → r.printParseErrors(); return
    │
    ▼
evaluator.Eval(program, r.env) → object.Object
    │
    ├── if nil or object.NULL  → return (print nothing)
    │
    ▼
fmt.Fprintln(r.out, result.Inspect())
```

Key decisions:
- Each line gets its **own fresh Lexer and Parser** — they have no state worth preserving
  between lines.
- The **environment is shared** — `r.env` carries bindings across all calls to `eval`.
- Parse errors short-circuit before evaluation — `evaluator.Eval` is never called on a
  program with parse errors.

---

### `printParseErrors`

```go
func (r *REPL) printParseErrors(errors []*parser.ParseError)
```

Prints a header line `"parse errors:"` followed by each error indented with a tab. Writes to
`r.out` (not `os.Stderr`) so tests can capture and assert on the error output.

Output format:
```
parse errors:
	parse error at '<token>': <message>
	parse error at '<token>': <message>
```

---

## `cmd/zen/main.go` — Entry Point Reference

`main.go` is intentionally minimal. It makes one decision — file mode or interactive mode —
and delegates everything else.

### `main`

```go
func main()
```

Checks `os.Args`:

| `os.Args` length | Mode | Action |
|-----------------|------|--------|
| `> 1` | File execution | `runFile(os.Args[1])` |
| `== 1` | Interactive REPL | `printWelcome()` then `repl.New(os.Stdin, os.Stdout).Start()` |

---

### `printWelcome`

```go
func printWelcome()
```

Prints the welcome banner to stdout. Only called in interactive mode — file execution is
silent on success. Output:

```
Welcome to Zen v0.1
Type 'exit' or 'quit' to leave

```

---

### `runFile`

```go
func runFile(filename string)
```

Executes a `.zen` source file non-interactively. Runs the same
`Lexer → Parser → Evaluator` pipeline as the REPL's `eval()` but:

- Reads the entire file at once rather than line by line.
- Writes errors to `os.Stderr` (not stdout) so they are distinguishable in shell pipelines.
- Calls `os.Exit(1)` on any error — file execution is all-or-nothing.
- Prints nothing on success — the script's output comes from explicit `print()` calls.

Error handling:

| Condition | Output | Exit code |
|-----------|--------|-----------|
| File not found / unreadable | `error: could not read file '<name>': <err>` to stderr | 1 |
| Parse errors | `parse errors:` + each error to stderr | 1 |
| Runtime error | `ERROR: <message>` to stderr | 1 |
| Success | nothing | 0 |

This exit code behaviour makes the binary usable in shell pipelines:

```bash
zen script.zen && echo "success" || echo "failed"
```

---

## Test Coverage — `repl_test.go`

### Test Helper — `runREPL`

```go
func runREPL(input string) string
```

Creates a fresh `REPL` with a `strings.NewReader` as input and a `bytes.Buffer` as output,
calls `Start()`, and returns the captured output as a string. Every test uses this helper —
no test ever touches a real terminal or file.

The pattern allows any multi-line session to be simulated by joining lines with `\n`:

```go
output := runREPL("let x = 10\nx + 5\n")
// output contains everything the REPL printed: prompts, results, errors
```

### Test Index

| Test | What it verifies |
|------|-----------------|
| `TestSingleExpression` | `5 + 3` → output contains `"8"` |
| `TestArithmeticPrecedence` | `2 + 3 * 4` → output contains `"14"` (multiplication binds tighter) |
| `TestBooleanExpression` | `5 > 3` → output contains `"true"` |
| `TestStringExpression` | `"hello" + " " + "world"` → output contains `"hello world"` |
| `TestVariablePersistenceAcrossLines` | `let x = 10` on line 1, `x + 5` on line 2 → output contains `"15"` |
| `TestMultipleVariablesPersist` | `let a = 3`, `let b = 7`, `a * b` → output contains `"21"` |
| `TestFunctionDefinitionAndCall` | Define `add` on line 1, call it on line 2 → output contains `"10"` |
| `TestClosurePersists` | `makeAdder(5)` returns closure; `addFive(3)` → output contains `"8"` |
| `TestNullNotPrintedForLetStatement` | `let x = 5` → output does NOT contain `"null"` or `"NULL"` |
| `TestNullNotPrintedForFunctionDefinition` | `let f = fn(x) { x }` → output does NOT contain `"null"` |
| `TestParseErrorIsReported` | `let = 5` → output contains `"parse errors:"` |
| `TestREPLContinuesAfterParseError` | Invalid line then `1 + 1` → output contains both the error and `"2"` |
| `TestRuntimeErrorIsDisplayed` | `5 + true` → output contains `"ERROR"` |
| `TestREPLContinuesAfterRuntimeError` | `5 + true` then `3 * 3` → output contains `"9"` |
| `TestUndefinedVariableError` | `foobar` → output contains `"identifier not found"` |
| `TestExitCommand` | `exit` → output contains `"Goodbye!"` |
| `TestQuitCommand` | `quit` → output contains `"Goodbye!"` |
| `TestExitStopsEvaluation` | `exit` then `99 + 1` → output does NOT contain `"100"` |
| `TestPromptIsPresent` | Single line input → output contains `"zen> "` |
| `TestPromptAppearsOnEveryLine` | Three lines → output contains at least 3 `"zen> "` strings |
| `TestArrayLiteral` | `[1, 2, 3]` → output contains `"[1, 2, 3]"` |
| `TestHashLiteral` | `{"x": 1}` → output contains both `"x"` and `"1"` |
| `TestBuiltinLen` | `len("hello")` → output contains `"5"` |
| `TestEOFExitsGracefully` | Empty input → `Start()` returns without panicking |

### Running Tests

```bash
# Run repl tests only
go test ./pkg/repl/...

# Verbose — see each test name
go test -v ./pkg/repl/...

# Run a single test
go test -v -run TestVariablePersistenceAcrossLines ./pkg/repl/...

# Run all packages
go test ./...
```

---

## Example Session

```
$ ./zen
Welcome to Zen v0.1
Type 'exit' or 'quit' to leave

zen> let x = 10
zen> let y = 20
zen> x + y
30
zen> let add = fn(a, b) { a + b }
zen> add(x, y)
30
zen> let factorial = fn(n) { if (n <= 1) { return 1 } return n * factorial(n - 1) }
zen> factorial(5)
120
zen> let arr = [1, 2, 3]
zen> push(arr, 4)
[1, 2, 3, 4]
zen> let = broken
parse errors:
	parse error at '=': no prefix parse function found for token type '='
zen> 5 + true
ERROR: type mismatch: INTEGER + BOOLEAN
zen> exit
Goodbye!
```

---

## Build & Run

```bash
# Build the zen binary
go build -o zen ./cmd/zen/main.go

# Start the interactive REPL
./zen

# Run a source file
./zen examples/fibonacci.zen

# Build with size optimisation
go build -ldflags="-s -w" -o zen ./cmd/zen/main.go
```