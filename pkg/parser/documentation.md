# Parser Package — Documentation

**Package:** `github.com/Vamshi-gande/zenlang/pkg/parser`  
**Location:** `pkg/parser/`  
**Files:** `parser.go`, `precedence.go`, `error.go`, `parser_test.go`

---

## Overview

The parser sits between the lexer and the evaluator. It consumes the flat token stream produced by the lexer and builds a structured AST that the evaluator can walk. Think of it as taking a list of words and arranging them into a grammatically structured diagram.

```
Lexer (token stream) → Parser → AST (consumed by evaluator)
```

The parser uses **Pratt parsing** (top-down operator precedence) for expressions. Instead of one giant switch statement, each token type has a small dedicated parse function registered to it. When the parser sees a token, it looks up and calls the right function.

---

## File Structure

```
pkg/parser/
├── parser.go        → Parser struct, all parse functions, entry point
├── precedence.go    → Precedence constants, lookup table, helper methods
├── error.go         → ParseError type, error recording helpers
└── parser_test.go   → Full test suite, end-to-end through the lexer
```

---

## precedence.go — Operator Precedence

### Why It Exists

When the parser sees `5 + 3 * 2`, it needs to know that `*` binds tighter than `+` — so the result should be `5 + (3 * 2)`, not `(5 + 3) * 2`. Precedence levels encode this as integers. A higher number means tighter binding.

### Precedence Constants

```
Level 1  LOWEST       — starting point, weakest possible binding
Level 2  EQUALS       — == != && || and compound assignments
Level 3  LESSGREATER  — < > <= >=
Level 4  SUM          — + -
Level 5  PRODUCT      — * /
Level 6  PREFIX       — -x  !x  ++x  --x
Level 7  CALL         — fn(args)
Level 8  INDEX        — arr[0]
```

### Precedence Lookup Table

Maps every infix or postfix token type to its level. When the Pratt loop needs to decide whether to keep consuming to the right, it calls `peekPrecedence()` and compares it to the current level.

| Token(s) | Level |
|---|---|
| `==` `!=` `&&` `\|\|` `+=` `-=` `*=` `/=` | EQUALS (2) |
| `<` `>` `<=` `>=` | LESSGREATER (3) |
| `+` `-` | SUM (4) |
| `*` `/` | PRODUCT (5) |
| `(` (call) | CALL (7) |
| `[` (index) | INDEX (8) |

`&&` and `||` sit at `EQUALS` — they bind loosely, at the same level as equality comparisons. Compound assignments (`+=` etc.) also sit at `EQUALS` since they are the weakest of all operators.

### Helper Methods

Both are methods on `*Parser` so they can read the parser's current state directly.

**`peekPrecedence()`** — returns the precedence of `peekToken`. Returns `LOWEST` if the token has no entry in the table. Used in the Pratt loop condition.

**`currentPrecedence()`** — returns the precedence of `currentToken`. Used by `parseInfixExpression` to pass the correct level when recursing right.

---

## error.go — Error Handling

### ParseError

```go
type ParseError struct {
    Token   token.Token  // the token where the problem occurred
    Message string       // what was expected vs what was found
}
```

Implements the standard `error` interface. `Error()` formats as:

```
parse error at '<literal>': <message>
```

For example: `parse error at '=': expected next token to be 'IDENT', got '=' instead`

### Error Collection Strategy

The parser **never stops at the first error**. All errors are appended to `p.errors` and returned together at the end via `Errors()`. This means a single parse pass reports every problem in the source, not just the first one.

### Error Recording Functions

**`peekError(expected token.TokenType)`** — called when `expectPeek` fails. Records that the next token was not what the grammar required. Message format: `expected next token to be 'X', got 'Y' instead`.

**`noPrefixParseFnError(t token.TokenType)`** — called inside `parseExpression` when no prefix function is registered for `currentToken`. This means the token cannot legally start an expression. Message format: `no prefix parse function found for token type 'X'`.

---

## parser.go — The Parser

### Internal State

```go
type Parser struct {
    lexer        *lexer.Lexer               // source of tokens
    currentToken token.Token                // token being examined right now
    peekToken    token.Token                // next token coming up
    errors       []*ParseError              // accumulated errors

    prefixParseFns map[token.TokenType]prefixParseFn
    infixParseFns  map[token.TokenType]infixParseFn
}
```

`currentToken` and `peekToken` give the parser a two-token window into the stream. One token of lookahead is enough to make every parsing decision in Zen.

The two function maps are the heart of Pratt parsing — each token type is mapped to the function responsible for parsing it.

### Construction — `NewParser(l *lexer.Lexer)`

Three things happen during construction:

1. `registerParseFunctions()` is called to populate both maps.
2. `nextToken()` is called once — this primes `peekToken`.
3. `nextToken()` is called again — this moves `peekToken` into `currentToken` and reads a new `peekToken`.

After construction, both `currentToken` and `peekToken` hold real tokens and the parser is ready to call `ParseProgram()`.

---

## Token Navigation

### `nextToken()`

Advances the two-token window. `peekToken` becomes `currentToken`, and a fresh token is fetched from the lexer into `peekToken`.

```
Before:  currentToken = 5     peekToken = +
After:   currentToken = +     peekToken = 3
```

### `currentTokenIs(t)` and `peekTokenIs(t)`

Simple boolean checks. Used throughout the parser to branch on what is being seen now vs what is coming next.

### `expectPeek(t token.TokenType) bool`

The most important helper in the entire parser. It enforces grammar rules one token at a time.

- If `peekToken` matches `t` → advance and return `true`
- If `peekToken` does not match → record a `peekError` and return `false`

Every grammar rule that requires a specific next token uses this. If it returns `false`, the calling parse function returns `nil` to signal failure without panicking.

---

## Entry Point — `ParseProgram()`

The only public parsing method. Creates the root `*ast.Program` and loops calling `parseStatement()` until `EOF`.

```
1. Create empty Program
2. While currentToken != EOF:
   a. parseStatement() → stmt
   b. If stmt != nil, append to Program.Statements
   c. nextToken()
3. Return Program
```

The `nextToken()` at the end of each loop iteration advances past the last token consumed by the statement parser, positioning `currentToken` on the first token of the next statement.

---

## Statement Parsing

### `parseStatement()`

Routes to the correct statement parser by looking at `currentToken.Type`.

| Token | Parser called |
|---|---|
| `LET` | `parseLetStatement()` |
| `RETURN` | `parseReturnStatement()` |
| `WHILE` | `parseWhileStatement()` |
| anything else | `parseExpressionStatement()` |

### `parseLetStatement()`

Grammar: `let <identifier> = <expression> [;]`

Uses `expectPeek` at each required token. If any step fails, returns `nil` immediately. The trailing semicolon is optional — the parser peeks for it and consumes it if present.

```
Token sequence consumed:
  LET  →  IDENT  →  ASSIGN  →  <expression>  →  [SEMICOLON]
```

### `parseReturnStatement()`

Grammar: `return <expression> [;]`

Advances past `return`, then calls `parseExpression(LOWEST)` for the value. Trailing semicolon is optional.

### `parseWhileStatement()`

Grammar: `while ( <condition> ) { <body> }`

Uses `expectPeek` to enforce the required `(`, `)`, and `{` tokens. The body is parsed by `parseBlockStatement()`.

```
Token sequence consumed:
  WHILE  →  LPAREN  →  <condition>  →  RPAREN  →  LBRACE  →  <block>
```

### `parseExpressionStatement()`

Wraps any expression that appears as a standalone statement. Calls `parseExpression(LOWEST)` — the `LOWEST` starting precedence means the expression parser greedily consumes as much as it can. Trailing semicolon is optional.

### `parseBlockStatement()`

Grammar: `{ <statement>* }`

Called with `currentToken` positioned on `{`. Advances past `{` then loops calling `parseStatement()` until it hits `}` or `EOF`. Used for function bodies and `if`/`else`/`while` branches.

---

## Expression Parsing — The Pratt Loop

`parseExpression(precedence int)` is the most important method in the parser. All expression parsing flows through it.

### The Algorithm

```
1. Look up prefix function for currentToken.Type
2. If none found → record noPrefixParseFnError, return nil
3. Call prefix function → left expression
4. Loop:
   a. If peekToken is SEMICOLON → stop
   b. If peekPrecedence() <= precedence → stop
   c. Look up infix function for peekToken.Type
   d. If none found → return left as-is
   e. nextToken()
   f. Call infix function with left → new left
5. Return left
```

### What the Precedence Parameter Controls

The `precedence` argument is the level the current parse context is operating at. The loop only continues consuming to the right when the next operator's precedence is **strictly higher** than the current level.

Calling with `LOWEST` (1) consumes the maximum — everything to the right until a semicolon or a closing delimiter.

Calling with `PRODUCT` (5) stops before consuming `+` or `-` (level 4), which is how `5 * 3 + 2` correctly becomes `(5 * 3) + 2` rather than `5 * (3 + 2)`.

### Worked Example — `5 + 3 * 2`

```
parseExpression(LOWEST)
  prefixFn for INT → left = IntegerLiteral(5)
  peekToken = +, peekPrecedence = SUM(4) > LOWEST(1) → continue
    nextToken() → currentToken = +
    infixFn for + → parseInfixExpression(left=5)
      operator = "+", precedence = SUM(4)
      nextToken() → currentToken = 3
      parseExpression(SUM)
        prefixFn for INT → left = IntegerLiteral(3)
        peekToken = *, peekPrecedence = PRODUCT(5) > SUM(4) → continue
          nextToken() → currentToken = *
          infixFn for * → parseInfixExpression(left=3)
            operator = "*", precedence = PRODUCT(5)
            nextToken() → currentToken = 2
            parseExpression(PRODUCT)
              prefixFn for INT → left = IntegerLiteral(2)
              peekToken = EOF, peekPrecedence = LOWEST(1) <= PRODUCT(5) → stop
            return InfixExpression(3, *, 2)
          peekToken = EOF, peekPrecedence = LOWEST(1) <= SUM(4) → stop
        return InfixExpression(3, *, 2)
      return InfixExpression(5, +, InfixExpression(3, *, 2))
  peekToken = EOF, peekPrecedence = LOWEST(1) <= LOWEST(1) → stop
return InfixExpression(5, +, InfixExpression(3, *, 2))
```

Result: `(5 + (3 * 2))` — correct.

---

## Prefix Parse Functions

Each is registered in `prefixParseFns` and called when its token type appears at the start of an expression.

### `parseIdentifier()`

Simplest of all. Wraps `currentToken` in an `*ast.Identifier` and returns it. No advancement — the Pratt loop advances before calling the next function.

### `parseIntegerLiteral()`

Reads `currentToken.Literal` and converts it to `int64` using `strconv.ParseInt`. Records a parse error and returns `nil` if conversion fails.

### `parseFloatLiteral()`

Same pattern as integer — converts `currentToken.Literal` to `float64` using `strconv.ParseFloat`. The lexer is responsible for producing a `FLOAT` token with the correct `"3.14"` literal.

### `parseStringLiteral()`

Wraps `currentToken.Literal` directly in `*ast.StringLiteral`. The lexer has already stripped the surrounding quotes.

### `parseBoolean()`

Returns a `*ast.BooleanLiteral`. Sets `Value` to `true` if `currentToken.Type == token.TRUE`, `false` otherwise. Handles both `TRUE` and `FALSE` tokens with one function.

### `parseNullLiteral()`

Wraps `currentToken` in a `*ast.NullLiteral`. No value field needed — there is only one null.

### `parsePrefixExpression()`

Handles: `!<expr>` `-<expr>` `++<expr>` `--<expr>`

Records the operator from `currentToken`, advances once, then calls `parseExpression(PREFIX)`. The `PREFIX` precedence level (6) is high, so the right operand binds tightly — `-a * b` parses as `(-a) * b`, not `-(a * b)`.

### `parseGroupedExpression()`

Handles `( <expr> )`. Advances past `(`, calls `parseExpression(LOWEST)` to parse the inner expression with a fresh precedence context, then uses `expectPeek(RPAREN)` to consume the closing `)`. The grouping itself creates no AST node — its only effect is resetting precedence to `LOWEST` inside the parentheses.

### `parseIfExpression()`

Handles: `if ( <condition> ) { <consequence> } [else { <alternative> }]`

Uses `expectPeek` to enforce each required delimiter. The `else` branch is optional — the parser peeks for `token.ELSE` after the consequence block closes.

```
Token sequence consumed:
  IF → LPAREN → <condition> → RPAREN → LBRACE → <consequence block>
  [ELSE → LBRACE → <alternative block>]
```

### `parseFunctionLiteral()`

Handles: `fn ( <params> ) { <body> }`

The keyword in source is `fn` but the token type is `token.FUNCTION` (as defined in the lexer's keyword map). Delegates parameter parsing to `parseFunctionParameters()`.

### `parseFunctionParameters()`

Parses the comma-separated identifier list between `(` and `)`. Handles zero parameters cleanly — peeks for `)` before trying to read any identifiers. Returns `nil` on error, which propagates up to the function literal as a failed parse.

### `parseArrayLiteral()`

Handles: `[ <expr>, <expr>, ... ]`

Delegates entirely to `parseExpressionList(token.RBRACKET)`.

### `parseHashLiteral()`

Handles: `{ <key> : <value>, ... }`

Loops until it sees `}`. On each iteration: parses key, uses `expectPeek(token.COLON)` to consume `:`, parses value, then either finds `,` to continue or `}` to stop. Both keys and values are full expressions.

---

## Infix Parse Functions

Each receives the already-parsed left expression as an argument and returns a combined expression.

### `parseInfixExpression(left)`

Handles all binary operators: `+` `-` `*` `/` `<` `>` `<=` `>=` `==` `!=` `&&` `||` `+=` `-=` `*=` `/=`

Records the operator and calls `currentPrecedence()` to capture the current level, advances past the operator, then calls `parseExpression(currentPrecedence)` to get the right operand. Passing `currentPrecedence` rather than `currentPrecedence - 1` makes all these operators **left-associative** — `1 + 2 + 3` parses as `(1 + 2) + 3`.

### `parseCallExpression(left)`

Handles: `<expr> ( <args> )`

`left` is the function being called — it can be an `Identifier` or even an inline `FunctionLiteral`. Delegates argument parsing to `parseExpressionList(token.RPAREN)`.

### `parseIndexExpression(left)`

Handles: `<expr> [ <index> ]`

`left` is the object being indexed. Advances past `[`, parses the index expression with `LOWEST` precedence, then uses `expectPeek(RBRACKET)` to consume `]`.

---

## Shared Helper — `parseExpressionList(end token.TokenType)`

Used by both `parseCallExpression` and `parseArrayLiteral` to avoid code duplication. Parses a comma-separated list of expressions terminated by the given closing token.

```
currentToken must be the opening delimiter on entry (LPAREN or LBRACKET).

1. If peekToken == end → empty list, advance, return []
2. nextToken() → first element
3. parseExpression(LOWEST) → append
4. While peekToken == COMMA:
   a. nextToken() past comma
   b. nextToken() to next element
   c. parseExpression(LOWEST) → append
5. expectPeek(end) → consume closing delimiter
6. Return list
```

---

## Function Registration Map

### Prefix Functions

| Token | Function | AST Node Produced |
|---|---|---|
| `IDENT` | `parseIdentifier` | `*ast.Identifier` |
| `INT` | `parseIntegerLiteral` | `*ast.IntegerLiteral` |
| `FLOAT` | `parseFloatLiteral` | `*ast.FloatLiteral` |
| `STRING` | `parseStringLiteral` | `*ast.StringLiteral` |
| `TRUE` `FALSE` | `parseBoolean` | `*ast.BooleanLiteral` |
| `NULL` | `parseNullLiteral` | `*ast.NullLiteral` |
| `BANG` `MINUS` `INC` `DEC` | `parsePrefixExpression` | `*ast.PrefixExpression` |
| `LPAREN` | `parseGroupedExpression` | (no node, resets precedence) |
| `IF` | `parseIfExpression` | `*ast.IfExpression` |
| `FUNCTION` | `parseFunctionLiteral` | `*ast.FunctionLiteral` |
| `LBRACKET` | `parseArrayLiteral` | `*ast.ArrayLiteral` |
| `LBRACE` | `parseHashLiteral` | `*ast.HashLiteral` |

### Infix Functions

| Token(s) | Function | AST Node Produced |
|---|---|---|
| `+` `-` `*` `/` `<` `>` `<=` `>=` `==` `!=` `&&` `\|\|` `+=` `-=` `*=` `/=` | `parseInfixExpression` | `*ast.InfixExpression` |
| `LPAREN` | `parseCallExpression` | `*ast.CallExpression` |
| `LBRACKET` | `parseIndexExpression` | `*ast.IndexExpression` |

Note that `LPAREN` and `LBRACKET` each appear in **both** maps — as prefix tokens they start array/hash literals and grouped expressions; as infix tokens they indicate a call or index operation on the preceding expression.

---

## Full Dependency Picture

```
parser.go
    ├── lexer.Lexer          pkg/lexer   — source of tokens
    ├── ast nodes            pkg/ast     — nodes being constructed
    ├── token constants      pkg/token   — token type names
    ├── precedence table     precedence.go
    └── ParseError           error.go
```

---

## Testing

All tests go through the full `lexer → parser → AST` pipeline. No tokens are hand-crafted. Every test parses a real source string and inspects the resulting AST.

### Test Helpers

| Helper | Purpose |
|---|---|
| `parse(t, input)` | Full pipeline, fails immediately on any parse error |
| `parseWithErrors(input)` | Full pipeline, returns parser so errors can be inspected directly |
| `checkParserErrors(t, p)` | Fails the test and prints all errors if any were collected |
| `requireStatements(t, program, n)` | Asserts exact statement count, returns statements |
| `requireExpressionStatement(t, stmt)` | Asserts stmt is ExpressionStatement, returns inner expression |
| `assertIntegerLiteral(t, expr, want)` | Type-asserts and value-checks an IntegerLiteral |
| `assertIdentifier(t, expr, want)` | Type-asserts and value-checks an Identifier |
| `assertBoolean(t, expr, want)` | Type-asserts and value-checks a BooleanLiteral |
| `assertLiteralExpression(t, expr, want)` | Dispatches to the above based on `want`'s Go type |
| `assertInfixExpression(t, expr, left, op, right)` | Full check of an InfixExpression node |

### Test Coverage

| Test | What It Verifies |
|---|---|
| `TestLetStatements` | `let x = 5`, `let y = true`, `let foo = y` |
| `TestLetStatementTokenLiteral` | TokenLiteral returns `"let"` |
| `TestReturnStatements` | `return 5`, `return true`, `return foobar` |
| `TestReturnInfixExpression` | `return x + y` produces InfixExpression |
| `TestIdentifierExpression` | `foobar` becomes Identifier node |
| `TestIntegerLiteralExpression` | `5` becomes IntegerLiteral(5) |
| `TestFloatLiteralExpression` | `3.14` becomes FloatLiteral(3.14) |
| `TestStringLiteralExpression` | `"hello"` becomes StringLiteral("hello") |
| `TestBooleanLiteralTrue/False` | `true`/`false` become BooleanLiteral nodes |
| `TestNullLiteral` | `null` becomes NullLiteral node |
| `TestPrefixExpressions` | `!true`, `!false`, `-5`, `--x`, `++x` |
| `TestInfixExpressions` | All 12 binary operators |
| `TestCompoundAssignmentOperators` | `+=` `-=` `*=` `/=` |
| `TestOperatorPrecedence` | 9 cases verified via `String()` output |
| `TestGroupedExpression` | `(5 + 3)` produces flat InfixExpression |
| `TestIfExpressionNoElse` | Condition, consequence present; alternative nil |
| `TestIfExpressionWithElse` | Both branches present and correct |
| `TestWhileStatement` | Condition and body parsed correctly |
| `TestFunctionLiteralNoParams` | `fn() { 5 }` — zero parameters |
| `TestFunctionLiteralTwoParams` | `fn(x, y) { x + y }` — params and body |
| `TestFunctionLiteralThreeParams` | Three params parsed in order |
| `TestFunctionLiteralTokenLiteral` | TokenLiteral returns `"fn"` |
| `TestCallExpressionNoArgs` | `add()` — zero arguments |
| `TestCallExpressionTwoArgs` | `add(1, 2 * 3)` — second arg is InfixExpression |
| `TestCallExpressionIIFE` | `fn(x) { x * 2 }(5)` — inline function literal call |
| `TestArrayLiteralEmpty` | `[]` — zero elements |
| `TestArrayLiteralThreeElements` | `[1, 2 * 2, 3 + 3]` — expressions as elements |
| `TestIndexExpression` | `arr[0]` |
| `TestIndexExpressionWithExpression` | `arr[1 + 2]` — expression as index |
| `TestHashLiteralEmpty` | `{}` — zero pairs |
| `TestHashLiteralIntegerKeys` | `{1: 2, 3: 4}` — two pairs |
| `TestHashLiteralBooleanKeys` | `{true: 1, false: 2}` |
| `TestMultipleStatements` | Three statements in sequence |
| `TestRecursiveFunction` | Full factorial function parses without error |
| `TestParserErrorMissingIdentifierInLet` | `let = 5` produces errors |
| `TestParserErrorMissingAssignInLet` | `let x 5` produces errors |
| `TestParserErrorMessageContent` | Error message is non-empty |
| `TestParserErrorsDoNotPanic` | 9 malformed inputs never panic |
| `TestParseErrorFormat` | `Error()` string format is correct |
| `TestPeekAndCurrentPrecedence` | `peekPrecedence` and `currentPrecedence` return SUM for `+` |
| `TestUnknownTokenHasLowestPrecedence` | Unknown token returns LOWEST |

Run all parser tests with:

```bash
go test ./pkg/parser/...
```

Run with verbose output to see each subtest:

```bash
go test -v ./pkg/parser/...
```

---

## Key Design Decisions

**Pratt parsing over a recursive descent switch.** A single `parseExpression` function with registered sub-functions handles all expression types. Adding a new operator or literal requires only registering one new function — the core loop never changes.

**Two-token lookahead window.** `currentToken` and `peekToken` are enough for every grammar decision in Zen. `peekTokenIs()` lets the parser look one step ahead to choose between branches without consuming.

**`expectPeek` as the grammar enforcer.** Every required token in a grammar rule is checked through `expectPeek`. It advances on success and records a targeted error on failure, keeping grammar enforcement in one place rather than scattered across parse functions.

**Errors accumulate, never stop.** Returning `nil` on a failed `expectPeek` and continuing rather than panicking means all errors in a program are reported in one pass. This is far more useful during development than stopping at the first problem.

**Semicolons are optional.** All statement parsers peek for a trailing semicolon and consume it if present, but never require it. This keeps the language feel lightweight.

**`LPAREN` and `LBRACKET` are in both maps.** As prefix tokens they open grouped expressions, array literals, and hash literals. As infix tokens they indicate a call or index operation on the expression to their left. The Pratt loop handles the ambiguity automatically based on position.