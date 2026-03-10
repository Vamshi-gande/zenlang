# `pkg/lexer` — Package Documentation

**Package:** `github.com/Vamshi-gande/zenlang/pkg/lexer`  
**Location:** `pkg/lexer/`  
**Files:** `input_reader.go`, `lexer.go`, `lexer_test.go`

---

## Overview

The `lexer` package is the front-end of the Zen language pipeline. It takes raw source code as a string and converts it into a flat stream of tokens that the parser consumes. It has no knowledge of grammar or meaning — it only knows how to recognise characters and group them into labelled chunks.

```
Source String  →  [ lexer package ]  →  stream of token.Token
```

The package is split across two implementation files, each with a distinct responsibility.

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

`position` is where you are right now. `readPosition` is always one step ahead. This lets the lexer peek at the next character without consuming it — essential for disambiguating `=` from `==`, `+` from `++` and `+=`, `&` from `&&`, and so on.

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

`currentChar` is a local mirror of `reader.CurrentChar()` kept in sync after every `advance()` call. This avoids calling `reader.CurrentChar()` repeatedly in tight loops.

**`NextToken()` flow**

```
called
  ↓
skipWhitespace()
  ↓
switch on currentChar
  ├── single char token          → emit token, advance once, return
  ├── two/three char token       → peek, emit correct token, advance, return
  ├── string literal (")         → readString(), return STRING token
  ├── letter or underscore       → readWhile(isIdentChar) → LookupIdentifier → return
  ├── digit                      → readNumber() → return INT or FLOAT token
  ├── null byte (0)              → return EOF token
  └── anything unrecognised      → return ILLEGAL token, advance
```

Multi-character tokens produced by `readWhile` and `readNumber` use early `return` statements — they do not fall through to the final `return tok` at the bottom. This is because `readWhile` already leaves `currentChar` positioned on the first non-matching character, so no further advance is needed.

---

## Token Recognition — Complete Reference

### Single Character Tokens

One character maps directly to one token. The lexer emits the token and advances once.

| Character | Token Type | Literal |
|---|---|---|
| `(` | `LPAREN` | `(` |
| `)` | `RPAREN` | `)` |
| `{` | `LBRACE` | `{` |
| `}` | `RBRACE` | `}` |
| `[` | `LBRACKET` | `[` |
| `]` | `RBRACKET` | `]` |
| `,` | `COMMA` | `,` |
| `;` | `SEMICOLON` | `;` |
| `:` | `COLON` | `:` |

---

### Two and Three Character Tokens

The lexer peeks at the next character to decide which token to emit. Every case ends with one final `advance()` that moves past the last character consumed.

**`=` — assignment or equality**

| Peek | Token | Literal |
|---|---|---|
| `=` | `EQ` | `==` |
| anything else | `ASSIGN` | `=` |

**`!` — logical NOT or not-equal**

| Peek | Token | Literal |
|---|---|---|
| `=` | `NOT_EQ` | `!=` |
| anything else | `BANG` | `!` |

**`<` — less-than or less-than-or-equal**

| Peek | Token | Literal |
|---|---|---|
| `=` | `LTE` | `<=` |
| anything else | `LT` | `<` |

**`>` — greater-than or greater-than-or-equal**

| Peek | Token | Literal |
|---|---|---|
| `=` | `GTE` | `>=` |
| anything else | `GT` | `>` |

**`+` — addition, increment, or compound assignment** *(three-way peek)*

| Peek | Token | Literal |
|---|---|---|
| `+` | `INC` | `++` |
| `=` | `PLUS_ASSIGN` | `+=` |
| anything else | `PLUS` | `+` |

**`-` — subtraction, decrement, or compound assignment** *(three-way peek)*

| Peek | Token | Literal |
|---|---|---|
| `-` | `DEC` | `--` |
| `=` | `MINUS_ASSIGN` | `-=` |
| anything else | `MINUS` | `-` |

**`*` — multiplication or compound assignment**

| Peek | Token | Literal |
|---|---|---|
| `=` | `ASTERISK_ASSIGN` | `*=` |
| anything else | `ASTERISK` | `*` |

**`/` — division or compound assignment**

| Peek | Token | Literal |
|---|---|---|
| `=` | `SLASH_ASSIGN` | `/=` |
| anything else | `SLASH` | `/` |

**`&` — logical AND or illegal**

| Peek | Token | Literal |
|---|---|---|
| `&` | `AND` | `&&` |
| anything else | `ILLEGAL` | `&` |

A single `&` has no meaning in Zen so it becomes `ILLEGAL`. Only `&&` is a valid operator.

**`|` — logical OR or illegal**

| Peek | Token | Literal |
|---|---|---|
| `|` | `OR` | `\|\|` |
| anything else | `ILLEGAL` | `|` |

Same reasoning — a bare `|` is not valid Zen.

---

### String Literals

When `"` is encountered, `readString()` is called.

**`readString()` algorithm:**

```
1. Advance past the opening "
2. Record start position
3. Consume characters until " or null byte (EOF)
4. Slice source[start:position] — the content without quotes
5. Advance past the closing "
6. Return the inner content as the literal
```

The returned `STRING` token's `Literal` field contains the string content without surrounding quotes. An unterminated string (EOF before closing `"`) stops safely at `0` rather than panicking.

---

### Number Literals — `readNumber()`

Numbers always start with a digit. The lexer calls `readNumber()` from the `default` branch when `isDigit(currentChar)` is true.

**Algorithm:**

```
1. Record start position
2. Consume digits with advance()
3. If currentChar == '.' AND PeekChar is a digit:
   a. Advance past '.'
   b. Consume remaining digits
   c. Return FLOAT token with literal e.g. "3.14"
4. Otherwise return INT token with literal e.g. "42"
```

The peek at the next character after `.` is critical — it distinguishes `3.14` (float) from `arr.method` (integer followed by member access). If the character after `.` is not a digit, the `.` is left unconsumed and the number is returned as `INT`.

---

### Identifiers and Keywords

When `isLetter(currentChar)` is true, the lexer reads the full identifier with `readWhile(isIdentChar)`, then calls `token.LookupIdentifier()` to check if the result is a reserved keyword.

**Identifier rules:**

The first character must satisfy `isLetter` — `a–z`, `A–Z`, or `_`. Every subsequent character is consumed by `isIdentChar` which additionally allows digits. This means:

- `counter_1` → valid identifier ✓
- `_private` → valid identifier ✓
- `myVar2` → valid identifier ✓
- `1abc` → `INT("1")` followed by `IDENT("abc")` — a digit cannot start an identifier

**Keywords** — `LookupIdentifier` returns a keyword token type when the literal matches exactly. Partial matches do not fire. `letter`, `ifelse`, `returned`, `truthy` all resolve to `IDENT`, not their embedded keywords.

| Keyword | Token Type |
|---|---|
| `let` | `LET` |
| `fn` | `FUNCTION` |
| `true` | `TRUE` |
| `false` | `FALSE` |
| `null` | `NULL` |
| `return` | `RETURN` |
| `if` | `IF` |
| `else` | `ELSE` |
| `while` | `WHILE` |

---

## Helper Functions

| Function | Signature | Purpose |
|---|---|---|
| `advance()` | `()` | Calls `reader.Advance()`, updates `currentChar` |
| `skipWhitespace()` | `()` | Consumes spaces `' '`, tabs `'\t'`, newlines `'\n'`, carriage returns `'\r'` |
| `readWhile()` | `(func(byte) bool) string` | Accumulates chars while condition holds; cursor lands on first non-matching char |
| `readNumber()` | `() token.Token` | Reads integer or float; returns the complete `token.Token` directly |
| `readString()` | `() string` | Reads between double quotes; returns inner content without quotes |
| `isLetter()` | `(byte) bool` | `a–z`, `A–Z`, underscore — valid first character of an identifier |
| `isIdentChar()` | `(byte) bool` | `isLetter` OR digit — valid continuation character of an identifier |
| `isDigit()` | `(byte) bool` | `0–9` |

---

## What Changed from the Original Lexer

The original lexer was missing several token categories. These were added to fix parser test failures:

| Addition | Reason |
|---|---|
| `case ':'` → `COLON` | Hash literals `{"key": val}` require a colon token; was producing `ILLEGAL` |
| `+` three-way peek for `++` and `+=` | `INC` and `PLUS_ASSIGN` were never produced |
| `-` three-way peek for `--` and `-=` | `DEC` and `MINUS_ASSIGN` were never produced |
| `*` peek for `*=` | `ASTERISK_ASSIGN` was never produced |
| `/` peek for `/=` | `SLASH_ASSIGN` was never produced |
| `case '&'` peek for `&&` | `AND` was never produced; `&` fell to `ILLEGAL` |
| `case '|'` peek for `\|\|` | `OR` was never produced; `|` fell to `ILLEGAL` |
| `readNumber()` with float detection | `3.14` was being split into `INT("3")` + `ILLEGAL(".")` + `INT("14")` |
| `case '"'` + `readString()` | String literals were entirely unhandled |

---

## Dependencies

```
lexer.go
    └── input_reader.go   (same package)
    └── pkg/token         (token types and LookupIdentifier)
```

`error.go` (if present) depends only on `fmt` from the standard library.

---

## Test Coverage

All tests drive the full pipeline: `NewLexer(input)` → repeated `NextToken()` calls → compare against expected token sequence.

**`runLexerTest(t, input, []expectedToken)`** — shared helper used across all tests. Calls `NextToken()` in a loop and on mismatch reports the exact index, expected type/literal, and received type/literal:

```
token[3]: expected type="INT" literal="10", got type="ILLEGAL" literal="1"
```

### Test Cases

| Test | What It Exercises |
|---|---|
| `TestSingleCharTokens` | Every single-character operator and delimiter in one input string, including `:` |
| `TestTwoCharTokens` | All two-char operators plus their single-char variants — verifies peek does not over-consume |
| `TestThreeWayPeekTokens` | `++`, `--`, `+=`, `-=`, `*=`, `/=` alongside bare `+`, `-`, `*`, `/` |
| `TestLogicalOperators` | `&&` and `\|\|` produce `AND` and `OR`; single `&` and `|` produce `ILLEGAL` |
| `TestWhitespaceHandling` | Same token sequence produced from compact, spaced, tabbed, and newline-separated input |
| `TestIntegerLiterals` | Single and multi-digit numbers are each one `INT` token |
| `TestFloatLiterals` | `3.14`, `0.5`, `100.001` produce `FLOAT` tokens; `3.` (no digits after dot) produces `INT` |
| `TestStringLiterals` | `"hello"`, `"hello world"` produce `STRING` tokens with quotes stripped |
| `TestIdentifiers` | Plain names, mixed case, underscores, digits after first char (`counter_1`) |
| `TestKeywords` | Every reserved word returns its keyword token type, not `IDENT` |
| `TestKeywordsInsideIdentifiers` | `letter`, `ifelse`, `returned`, `truthy` — partial matches must not fire |
| `TestLetStatement` | Full `let x = 10;` sequence end-to-end |
| `TestFunctionDefinition` | Full `let add = fn(a, b) { return a + b; }` sequence |
| `TestConditional` | Full `if (x == y) { return true; } else { return false; }` sequence |
| `TestHashLiteral` | `{"name": "Alice"}` — verifies `LBRACE`, `STRING`, `COLON`, `STRING`, `RBRACE` sequence |
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