# `pkg/token` — Package Documentation

## Overview

The `token` package defines the vocabulary of the Zen language. It is a pure data package — no parsing, no execution logic. Every other package that needs to talk about a token imports this one.

It provides three things:

- The `TokenType` type and the `Token` struct
- Every token type constant the lexer can produce
- `LookupIdentifier` — the one function that distinguishes keywords from plain identifiers

---

## Types

### `TokenType`

```go
type TokenType string
```

A named string type. Using a distinct type rather than raw `string` means the compiler catches mistakes where a token type is expected but a plain string is passed.

### `Token`

```go
type Token struct {
    Type    TokenType
    Literal string
}
```

`Type` is the category of the token — what kind of thing it is.  
`Literal` is the exact text from the source that produced it.

Every `NextToken()` call from the lexer returns one of these.

---

## Token Type Constants

### Special

| Constant | Value | Meaning |
|---|---|---|
| `ILLEGAL` | `"ILLEGAL"` | Character the lexer could not recognise |
| `EOF` | `"EOF"` | End of the source string |

### Literals

| Constant | Value | Meaning |
|---|---|---|
| `IDENT` | `"IDENT"` | User-defined identifier, e.g. `x`, `myVar`, `counter_1` |
| `INT` | `"INT"` | Integer literal, e.g. `5`, `123` |
| `FLOAT` | `"FLOAT"` | Float literal — defined, not yet lexed in Phase 1 |
| `STRING` | `"STRING"` | String literal — defined, not yet lexed in Phase 1 |

### Keywords

These are reserved words. The lexer produces them via `LookupIdentifier` — the source text `"let"` produces a `Token{Type: LET, Literal: "let"}`, not an `IDENT`.

| Constant | Source word | Phase introduced |
|---|---|---|
| `LET` | `let` | Phase 1 |
| `FUNCTION` | — | Phase 1 (internal alias) |
| `FN` | `fn` | Phase 1 |
| `TRUE` | `true` | Phase 2 |
| `FALSE` | `false` | Phase 2 |
| `NULL` | `null` | Phase 2 |
| `RETURN` | `return` | Phase 3 |
| `IF` | `if` | Phase 2 |
| `ELSE` | `else` | Phase 2 |
| `WHILE` | `while` | Phase 2 |

> `FN` and `FUNCTION` share the same string value `"FUNCTION"`. Use `FN` everywhere in the codebase — it matches the Zen keyword `fn` and is less verbose.

### Operators

| Constant | Value | Description |
|---|---|---|
| `ASSIGN` | `=` | Assignment |
| `PLUS` | `+` | Addition |
| `MINUS` | `-` | Subtraction |
| `BANG` | `!` | Logical NOT |
| `ASTERISK` | `*` | Multiplication |
| `SLASH` | `/` | Division |
| `LT` | `<` | Less than |
| `GT` | `>` | Greater than |
| `EQ` | `==` | Equality |
| `NOT_EQ` | `!=` | Inequality |
| `LTE` | `<=` | Less than or equal |
| `GTE` | `>=` | Greater than or equal |
| `AND` | `&&` | Logical AND — defined, not yet lexed in Phase 1 |
| `OR` | `\|\|` | Logical OR — defined, not yet lexed in Phase 1 |
| `INC` | `++` | Increment — defined, not yet lexed in Phase 1 |
| `DEC` | `--` | Decrement — defined, not yet lexed in Phase 1 |
| `PLUS_ASSIGN` | `+=` | Add-assign — defined, not yet lexed in Phase 1 |
| `MINUS_ASSIGN` | `-=` | Subtract-assign — defined, not yet lexed in Phase 1 |
| `ASTERISK_ASSIGN` | `*=` | Multiply-assign — defined, not yet lexed in Phase 1 |
| `SLASH_ASSIGN` | `/=` | Divide-assign — defined, not yet lexed in Phase 1 |

### Delimiters

| Constant | Value |
|---|---|
| `COMMA` | `,` |
| `SEMICOLON` | `;` |
| `COLON` | `:` |
| `LPAREN` | `(` |
| `RPAREN` | `)` |
| `LBRACE` | `{` |
| `RBRACE` | `}` |
| `LBRACKET` | `[` |
| `RBRACKET` | `]` |

---

## `LookupIdentifier`

```go
func LookupIdentifier(ident string) TokenType
```

The lexer calls this after reading a full word from the source. It checks the word against the `keywords` map and returns the keyword token type if found, or `IDENT` if not.

**Keywords map**

```
"let"    → LET
"fn"     → FUNCTION  (FN)
"true"   → TRUE
"false"  → FALSE
"null"   → NULL
"return" → RETURN
"if"     → IF
"else"   → ELSE
"while"  → WHILE
```

**Case sensitivity:** The map keys are all lowercase. `"let"` matches, `"Let"` and `"LET"` do not — they come back as `IDENT`. Zen is a case-sensitive language.

**Exact match only:** `"letter"` does not match `"let"`. The full identifier string is looked up, not a prefix of it.

---

## Dependencies

```
token.go
    └── no imports (pure data + one map lookup)
```

This package intentionally has zero dependencies. It sits at the bottom of the import graph — everything else can import it without risk of circular imports.

---

## Test File — `token_test.go`

Tests live in `package token` (same package, so unexported symbols are accessible if needed).

**Test cases**

| Test | What it exercises |
|---|---|
| `TestLookupIdentifier_Keywords` | Every keyword in the map returns its correct `TokenType` |
| `TestLookupIdentifier_Identifiers` | Plain words return `IDENT`; includes tricky cases like `letter`, `ifelse`, `returned`, `truthy` which contain keywords as substrings |
| `TestLookupIdentifier_CaseSensitive` | `Let`, `LET`, `Fn`, `FN`, `IF`, `True` etc. all return `IDENT` — not their keyword equivalents |

**Running the tests**

```bash
# Run all token tests
go test ./pkg/token/...

# Run with verbose output
go test -v ./pkg/token/...

# Run a specific test
go test -v -run TestLookupIdentifier_CaseSensitive ./pkg/token/...
```