# AST Package — Documentation

**Package:** `github.com/Vamshi-gande/zenlang/pkg/ast`  
**Location:** `pkg/ast/`  
**Files:** `ast.go`, `statements.go`, `expressions.go`, `ast_test.go`

---

## Overview

The AST package is a pure data definition layer. It defines the node types that represent every meaningful construct in Zen source code. It contains no parsing logic and no execution logic — just clean data structures that the parser populates and the evaluator later walks.

```
Tokens (from lexer) → Parser → AST Nodes → Evaluator
                                    ↑
                             defined here
```

Every node stores the `token.Token` it was built from. This is what makes `TokenLiteral()` work and what will later power accurate error messages with line information.

---

## File Structure

```
pkg/ast/
├── ast.go           → Core interfaces and Program root node
├── statements.go    → Statement node types
├── expressions.go   → Expression node types
└── ast_test.go      → Tests for all String() methods
```

---

## ast.go — Core Interfaces

### Node

The base interface every single AST node must satisfy, whether it is a statement, expression, or the root program.

| Method | Returns | Purpose |
|---|---|---|
| `TokenLiteral()` | `string` | The literal value of the token that originated this node. Used in tests and debugging. |
| `String()` | `string` | A string representation of the node and all its children. Used to print the full AST. |

### Statement

Embeds `Node`. Represents constructs that perform an action but do not produce a value you can use inside another expression. The marker method `statementNode()` is unexported — its presence is enough to satisfy the interface, it has no implementation.

```
let x = 5        → statement (you cannot write: let y = let x = 5)
return x         → statement
while x { ... }  → statement
```

### Expression

Embeds `Node`. Represents constructs that produce a value. The marker method `expressionNode()` works the same way as `statementNode()`.

```
5 + 3            → expression
factorial(n)     → expression
if (a > b) { a } else { b }  → expression
```

### Program

The root node of every AST. When the parser finishes, it returns a single `Program`. Everything else hangs off it.

```go
type Program struct {
    Statements []Statement
}
```

`TokenLiteral()` returns the literal of the first statement's token if one exists, otherwise an empty string. `String()` concatenates the `String()` output of every statement in order.

---

## statements.go — Statement Node Types

### LetStatement

Represents a variable binding: `let x = 5`

```go
type LetStatement struct {
    Token token.Token  // token.LET  —  literal "let"
    Name  *Identifier  // the variable being bound
    Value Expression   // the right-hand side
}
```

`String()` output: `let x = 5;`

`Value` may be `nil` during partial parsing — `String()` handles this without panicking.

---

### ReturnStatement

Represents an early return from a function: `return x + 1`

```go
type ReturnStatement struct {
    Token       token.Token  // token.RETURN  —  literal "return"
    ReturnValue Expression
}
```

`String()` output: `return x + 1;`

---

### ExpressionStatement

Wraps a bare expression that appears on its own line. This is the most common statement type in practice — almost every expression-only line like `factorial(n)` or `x + 1` becomes an `ExpressionStatement`.

```go
type ExpressionStatement struct {
    Token      token.Token  // first token of the expression
    Expression Expression
}
```

`String()` delegates directly to `Expression.String()`. Returns an empty string if `Expression` is `nil`.

---

### BlockStatement

Represents a `{ ... }` block — a sequence of statements wrapped in braces. Used as the body of functions and the branches of `if`/`else`.

```go
type BlockStatement struct {
    Token      token.Token  // token.LBRACE  —  literal "{"
    Statements []Statement
}
```

`String()` concatenates all inner statements. Note that `BlockStatement` is a `Statement` that contains other statements, not an expression, even though it appears inside expressions like `IfExpression` and `FunctionLiteral`.

---

### WhileStatement

Represents a while loop: `while (x > 0) { x = x - 1 }`

```go
type WhileStatement struct {
    Token     token.Token     // token.WHILE  —  literal "while"
    Condition Expression
    Body      *BlockStatement
}
```

`String()` output: `while<condition> <body>`

The `WHILE` token type corresponds to the `"while"` keyword in the lexer's keyword map.

---

## expressions.go — Expression Node Types

### Identifier

Represents a variable name used as a value in an expression: `x`, `counter`, `myFunc`.

```go
type Identifier struct {
    Token token.Token  // token.IDENT
    Value string       // the name itself, e.g. "x"
}
```

Also used to hold parameter names inside `FunctionLiteral.Parameters`. The same struct covers both roles.

`String()` returns `Value` directly.

---

### IntegerLiteral

Represents a whole number value: `5`, `100`, `42`

```go
type IntegerLiteral struct {
    Token token.Token  // token.INT
    Value int64
}
```

`Value` is stored as `int64` after the parser converts the token's string literal.

---

### FloatLiteral

Represents a floating point value: `3.14`, `0.5`

```go
type FloatLiteral struct {
    Token token.Token  // token.FLOAT
    Value float64
}
```

`Value` is stored as `float64` after the parser converts the token's string literal.

---

### StringLiteral

Represents a quoted string: `"hello"`, `"Alice"`

```go
type StringLiteral struct {
    Token token.Token  // token.STRING
    Value string
}
```

---

### BooleanLiteral

Represents `true` or `false`.

```go
type BooleanLiteral struct {
    Token token.Token  // token.TRUE or token.FALSE
    Value bool
}
```

`String()` returns the token's literal, which will be either `"true"` or `"false"`.

---

### NullLiteral

Represents the `null` keyword.

```go
type NullLiteral struct {
    Token token.Token  // token.NULL
}
```

`String()` always returns the fixed string `"null"`. No `Value` field is needed since there is only one possible null value.

---

### PrefixExpression

Represents a prefix operator applied to a single expression.

```go
type PrefixExpression struct {
    Token    token.Token  // the prefix token
    Operator string       // "!", "-", "++", "--"
    Right    Expression
}
```

| Operator | Token | Example |
|---|---|---|
| `!` | `token.BANG` | `!true` |
| `-` | `token.MINUS` | `-5` |
| `++` | `token.INC` | `++x` |
| `--` | `token.DEC` | `--x` |

`String()` output wraps in parentheses: `(!true)`, `(-5)`, `(++x)`

---

### InfixExpression

Represents a binary operation with a left operand, an operator, and a right operand.

```go
type InfixExpression struct {
    Token    token.Token  // the operator token
    Left     Expression
    Operator string
    Right    Expression
}
```

Covers all binary operators defined in the token package:

| Category | Operators |
|---|---|
| Arithmetic | `+`, `-`, `*`, `/` |
| Comparison | `<`, `>`, `<=`, `>=`, `==`, `!=` |
| Logical | `&&`, `\|\|` |
| Compound assignment | `+=`, `-=`, `*=`, `/=` |

`String()` output wraps in parentheses: `(5 + 3)`, `(x == y)`, `(a && b)`, `(x += 1)`

---

### IfExpression

Represents an `if`/`else` construct. In Zen, `if` is an expression — it produces a value, so `let x = if (a > b) { a } else { b }` is valid.

```go
type IfExpression struct {
    Token       token.Token      // token.IF  —  literal "if"
    Condition   Expression
    Consequence *BlockStatement
    Alternative *BlockStatement  // nil when there is no else branch
}
```

`Alternative` is `nil` when no `else` branch is present. The evaluator checks for this before executing the alternative.

---

### FunctionLiteral

Represents a function definition: `fn(a, b) { return a + b }`

```go
type FunctionLiteral struct {
    Token      token.Token    // token.FUNCTION  —  literal "fn"
    Parameters []*Identifier
    Body       *BlockStatement
}
```

The source keyword is `fn` but the token type is `token.FUNCTION` — this is defined in the lexer's keyword map as `"fn" → FUNCTION`.

Functions are first-class values in Zen, so `FunctionLiteral` is an `Expression`. This means `let add = fn(a, b) { a + b }` works naturally.

`String()` output: `fn(a, b)<body statements>`

---

### CallExpression

Represents a function call: `add(1, 2)`, `factorial(n)`

```go
type CallExpression struct {
    Token     token.Token   // token.LPAREN  —  the opening "("
    Function  Expression    // Identifier or FunctionLiteral
    Arguments []Expression
}
```

`Function` is itself an `Expression`, not just an identifier. This allows calling an inline function literal directly: `fn(x) { x * 2 }(5)`.

`String()` output: `add(1, 2)`

---

### IndexExpression

Represents index access on an array or hash map: `arr[0]`, `person["name"]`

```go
type IndexExpression struct {
    Token token.Token  // token.LBRACKET  —  literal "["
    Left  Expression
    Index Expression
}
```

Both `Left` and `Index` are expressions, so chained access like `matrix[i][j]` will parse as nested `IndexExpression` nodes.

`String()` output wraps in parentheses: `(arr[0])`, `(person[name])`

---

### ArrayLiteral

Represents an array literal: `[1, 2, 3]`

```go
type ArrayLiteral struct {
    Token    token.Token   // token.LBRACKET  —  literal "["
    Elements []Expression
}
```

Elements are a slice of expressions, so `[1 + 2, factorial(3)]` is valid.

`String()` output: `[1, 2, 3]`

---

### HashLiteral

Represents a hash map literal: `{"name": "Alice", "age": 30}`

```go
type HashLiteral struct {
    Token token.Token               // token.LBRACE  —  literal "{"
    Pairs map[Expression]Expression
}
```

Both keys and values are expressions. The colon separator between key and value corresponds to `token.COLON` in the token package. The parser is responsible for consuming it — the node itself only stores the resulting pairs.

`String()` output: `{name:Alice}`

---

## How the Three Files Relate

```
ast.go
  └── Node, Statement, Expression interfaces
  └── Program (root node)
        │
        ├── statements.go
        │     LetStatement
        │     ReturnStatement
        │     ExpressionStatement
        │     BlockStatement
        │     WhileStatement
        │
        └── expressions.go
              Identifier
              IntegerLiteral, FloatLiteral, StringLiteral
              BooleanLiteral, NullLiteral
              PrefixExpression, InfixExpression
              IfExpression
              FunctionLiteral, CallExpression
              IndexExpression, ArrayLiteral, HashLiteral
```

---

## Testing

`ast_test.go` tests every node type by manually constructing nodes — no parser is involved. The only thing being tested is `String()`, since that method is the primary tool for verifying correct tree construction when the parser is built.

Each test follows the same pattern: construct a node with real `token.Token` values from your token package, call `String()`, and assert the output.

| Test | Verifies |
|---|---|
| `TestProgramTokenLiteralEmpty` | Empty program returns `""` |
| `TestProgramTokenLiteralNonEmpty` | Non-empty program returns first token literal |
| `TestProgramStringConcatenatesStatements` | Multiple statements are joined correctly |
| `TestLetStatementString` | `let x = anotherVar;` |
| `TestLetStatementNilValue` | Nil value does not panic |
| `TestReturnStatementString` | `return x;` |
| `TestExpressionStatementString` | Delegates to inner expression |
| `TestExpressionStatementNilExpression` | Nil expression returns `""` |
| `TestBlockStatementString` | Inner statements are concatenated |
| `TestWhileStatementString` | Contains `"while"` |
| `TestIntegerLiteralString` | `"42"` |
| `TestFloatLiteralString` | `"3.14"` |
| `TestStringLiteralString` | `"hello"` |
| `TestBooleanLiteralTrue/False` | `"true"` / `"false"` |
| `TestNullLiteralString` | `"null"` |
| `TestIdentifierString` | Returns `Value` field |
| `TestPrefixExpressionBang` | `(!true)` |
| `TestPrefixExpressionMinus` | `(-5)` |
| `TestPrefixExpressionIncrement` | `(++x)` — uses `token.INC` |
| `TestInfixExpressionPlus` | `(5 + 3)` |
| `TestInfixExpressionEquality` | `(x == y)` — uses `token.EQ` |
| `TestInfixExpressionAnd` | `(a && b)` — uses `token.AND` |
| `TestInfixExpressionPlusAssign` | `(x += 1)` — uses `token.PLUS_ASSIGN` |
| `TestIfExpressionNoElse` | Contains `"if"`, does not contain `"else"` |
| `TestIfExpressionWithElse` | Contains `"else"` |
| `TestFunctionLiteralString` | Contains `"fn"` and parameter names |
| `TestCallExpressionString` | `add(1, 2)` |
| `TestArrayLiteralString` | `[1, 2, 3]` |
| `TestIndexExpressionString` | `(arr[0])` |
| `TestHashLiteralString` | Contains `{`, `}`, and `:` |

Run all tests with:

```bash
go test ./pkg/ast/...
```

---

## Design Decisions

**`if` is an expression, not a statement.** This means `let x = if (a > b) { a } else { b }` is valid syntax. It keeps the language consistent — everything that produces a value is an expression.

**Functions are expressions.** `FunctionLiteral` implements `Expression`, making functions first-class values. `let add = fn(a, b) { a + b }` works naturally without any special syntax.

**Every node stores its originating token.** The `Token` field on every node is what powers `TokenLiteral()` and will later enable error messages that include source location information.

**Marker methods are unexported.** `statementNode()` and `expressionNode()` have no implementation — their presence in a type is enough to satisfy the interface in Go's type system. Being unexported means only types within this package can implement `Statement` and `Expression`, preventing accidental misuse from outside packages.

**`BlockStatement` is a `Statement`.** Even though `BlockStatement` appears as a field inside expressions like `IfExpression` and `FunctionLiteral`, it implements `Statement`. This is correct — a block is a sequence of statements, and the last expression statement in a block produces the block's value at evaluation time.