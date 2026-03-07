# `pkg/lexer` — Package Documentation

## Overview

The `lexer` package is the front-end of the Zen language pipeline. It takes raw source code as a string and converts it into a flat stream of tokens that the parser consumes. It has no knowledge of grammar or meaning — it only knows how to recognise characters and group them into labelled chunks.

```
Source String  →  [ lexer package ]  →  stream of token.Token
```

The package is split across three files, each with a distinct responsibility.

---

## Files

### `input_reader.go`

The `InputReader` is a low-level cursor over the source string. It is the only part of the package that touches raw string indexing. Everything above it goes through its methods.

**Why two positions?**

```
source:       l  e  t     x
index:        0  1  2  3  4
              ↑        ↑
          position   readPosition
```

`position` is where you are right now. `readPosition` is always one step ahead. This lets the lexer peek at the next character without consuming it — essential for disambiguating `=` from `==`, `!` from `!=`, and so on.

**Key design detail — priming.** The constructor calls `Advance()` once immediately so that after `NewInputReader()` returns, `position` is already at index `0` and `readPosition` is at index `1`. The reader is ready to use without any further setup.

**Methods**

| Method | Returns | Description |
|---|---|---|
| `CurrentChar()` | `byte` | Character at `position`. Returns `0` at end of input. |
| `PeekChar()` | `byte` | Character at `readPosition` without advancing. Returns `0` at end. |
| `Advance()` | `byte` | Copies `readPosition` into `position`, increments `readPosition`, returns the new current char. |
| `IsAtEnd()` | `bool` | True when `position` has moved past the last valid index. |

**Edge cases handled:** empty string, single character source, reading past end — all return `0` safely rather than panicking.

---

### `lexer.go`

The `Lexer` sits on top of `InputReader` and applies language rules to the character stream. The parser only ever calls one method on it: `NextToken()`.

**Struct**

```go
type Lexer struct {
    reader      *InputReader
    currentChar byte
}
```

`currentChar` is a local mirror of `reader.CurrentChar()` kept in sync after every `advance()` call.

**`NextToken()` flow**

```
called
  ↓
skipWhitespace()
  ↓
switch on currentChar
  ├── single char token  → emit, advance once
  ├── two char token     → peek, emit, advance
  ├── letter / _         → readWhile(isIdentChar) → LookupIdentifier
  ├── digit              → readWhile(isDigit) → INT token
  ├── null byte (0)      → EOF token
  └── anything else      → ILLEGAL token
  ↓
return Token
```

**Token recognition categories**

Single character tokens — one character maps to one token, advance and return:
`+` `-` `*` `/` `(` `)` `{` `}` `[` `]` `,` `;`

Two character tokens — peek at the next char to decide:

| Seen | Peek | Token |
|---|---|---|
| `=` | `=` | `EQ` (`==`) |
| `=` | anything else | `ASSIGN` (`=`) |
| `!` | `=` | `NOT_EQ` (`!=`) |
| `!` | anything else | `BANG` (`!`) |
| `<` | `=` | `LTE` (`<=`) |
| `<` | anything else | `LT` (`<`) |
| `>` | `=` | `GTE` (`>=`) |
| `>` | anything else | `GT` (`>`) |

Multi-character tokens — `readWhile` accumulates the full string, then returns early without an extra advance because `readWhile` already leaves the cursor on the first non-matching character.

**Identifier rules**

The first character must satisfy `isLetter` (a–z, A–Z, `_`). Every subsequent character is consumed by `isIdentChar` which also allows digits. This means:

- `counter_1` → valid identifier ✓
- `_private` → valid identifier ✓  
- `1abc` → `INT("1")` then `IDENT("abc")` — digit cannot start an identifier

**Helper functions**

| Function | Signature | Purpose |
|---|---|---|
| `advance()` | `()` | Calls `reader.Advance()`, updates `currentChar` |
| `skipWhitespace()` | `()` | Consumes spaces, tabs, `\n`, `\r` |
| `readWhile()` | `(func(byte) bool) string` | Accumulates chars while condition holds |
| `isLetter()` | `(byte) bool` | a–z, A–Z, underscore |
| `isIdentChar()` | `(byte) bool` | `isLetter` OR digit — for identifier continuation |
| `isDigit()` | `(byte) bool` | 0–9 |

---

### `error.go`

Defines `LexerError`, a position-aware error type for when the lexer encounters an unrecognised character. It implements Go's standard `error` interface so it can be returned anywhere a normal error is expected.

**Types**

```go
type Position struct {
    Line   int
    Column int
}

type LexerError struct {
    Position Position
    Message  string
    Char     byte
}
```

**`Error() string`** formats the error as:
```
lexer error at line 3, column 12: unexpected character '@'
```

**Current state:** The infrastructure is defined and ready. Line and column tracking will be wired into `InputReader` during Phase 8 polish — `InputReader` will need to count newlines as it advances. Until then, `NewLexerError` accepts line and column as parameters so the call site controls the values.

---

## Dependencies

```
lexer.go
    └── input_reader.go   (same package)
    └── pkg/token         (token types and LookupIdentifier)

error.go
    └── fmt (standard library only)
```

---

## Test File — `lexer_test.go`

All tests live in `package lexer` and use a shared helper `runLexerTest` to avoid repetition.

**`runLexerTest(t, input, []expectedToken)`**

Drives a full token sequence comparison. Calls `NextToken()` in a loop and on failure reports the exact index, expected type/literal, and received type/literal:

```
token[3]: expected type="INT" literal="10", got type="ILLEGAL" literal="1"
```

**Test cases**

| Test | What it exercises |
|---|---|
| `TestSingleCharTokens` | Every single-character operator and delimiter in one input string |
| `TestTwoCharTokens` | All two-char operators plus their single-char variants — verifies peek does not over-consume |
| `TestWhitespaceHandling` | Same token sequence produced from compact, spaced, tabbed, and newline-separated input |
| `TestIntegerLiterals` | Single and multi-digit numbers are each one `INT` token |
| `TestIdentifiers` | Plain names, mixed case, underscores, digits after first char (`counter_1`) |
| `TestKeywords` | Every reserved word returns its keyword token type, not `IDENT` |
| `TestKeywordsInsideIdentifiers` | `letter`, `ifelse`, `returned`, `truthy` — partial matches must not fire |
| `TestLetStatement` | Full `let x = 10;` sequence end-to-end |
| `TestFunctionDefinition` | Full `let add = fn(a, b) { return a + b; }` sequence |
| `TestConditional` | Full `if (x == y) { return true; } else { return false; }` sequence |
| `TestEOFHandling` | Empty string gives immediate EOF; repeated calls after last token return EOF without panic |
| `TestIllegalCharacters` | `@` and `$` produce `ILLEGAL` tokens; lexer recovers and continues tokenising after them |

**Running the tests**

```bash
# Run all lexer tests
go test ./pkg/lexer/...

# Run with verbose output
go test -v ./pkg/lexer/...

# Run a single test by name
go test -v -run TestLetStatement ./pkg/lexer/...

# Run all tests in the project
go test ./...
```

**Known fix applied during testing:** `TestIdentifiers` initially failed on `counter_1` because `readWhile(isLetter)` stopped at the digit. Fixed by introducing `isIdentChar` and switching identifier continuation to use it instead of `isLetter`.