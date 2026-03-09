package ast

import (
	"bytes"
	"strings"

	"github.com/Vamshi-gande/zenlang/pkg/token"
)

// Identifier represents a variable name used as an expression, e.g. x, counter.
// Also used to hold parameter names in function literals.
// Token type: token.IDENT
type Identifier struct {
	Token token.Token // the IDENT token
	Value string      // the name itself
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents a raw integer value, e.g. 5 or 100.
// Token type: token.INT
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a floating point value, e.g. 3.14.
// Token type: token.FLOAT
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a quoted string value, e.g. "hello".
// Token type: token.STRING
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// BooleanLiteral represents true or false.
// Token types: token.TRUE / token.FALSE
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// NullLiteral represents the null keyword.
// Token type: token.NULL
type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return "null" }

// PrefixExpression represents a prefix operator applied to an expression.
// e.g.  !true   -5   ++x   --x
// Token types: token.BANG, token.MINUS, token.INC, token.DEC
type PrefixExpression struct {
	Token    token.Token // the prefix token
	Operator string      // "!", "-", "++", "--"
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression represents a binary operation.
// e.g.  5 + 3   x == y   a && b   x += 1
// Token types: token.PLUS, token.MINUS, token.ASTERISK, token.SLASH,
//
//	token.LT, token.GT, token.LTE, token.GTE,
//	token.EQ, token.NOT_EQ, token.AND, token.OR,
//	token.PLUS_ASSIGN, token.MINUS_ASSIGN,
//	token.ASTERISK_ASSIGN, token.SLASH_ASSIGN
type InfixExpression struct {
	Token    token.Token // the operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// IfExpression represents an if/else construct.
// In Zen, if is an expression — it produces a value.
// Token type: token.IF
// Alternative is nil when there is no else branch.
type IfExpression struct {
	Token       token.Token // the IF token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement // nil if no else branch
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

// FunctionLiteral represents a function definition.
// e.g.  fn(a, b) { return a + b }
// The keyword is "fn" but the token type is token.FUNCTION (see your keywords map).
type FunctionLiteral struct {
	Token      token.Token // the FUNCTION token  (literal "fn")
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := make([]string, len(fl.Parameters))
	for i, p := range fl.Parameters {
		params[i] = p.String()
	}
	out.WriteString(fl.TokenLiteral()) // "fn"
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(fl.Body.String())
	return out.String()
}

// CallExpression represents a function call.
// e.g.  add(1, 2)   factorial(n)
// Function is itself an Expression — it can be an Identifier or an inline FunctionLiteral.
// Token type: token.LPAREN  (the opening parenthesis of the call)
type CallExpression struct {
	Token     token.Token // the ( token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := make([]string, len(ce.Arguments))
	for i, a := range ce.Arguments {
		args[i] = a.String()
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// IndexExpression represents index access on arrays or hash maps.
// e.g.  arr[0]   person["name"]
// Token type: token.LBRACKET
type IndexExpression struct {
	Token token.Token // the [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

// ArrayLiteral represents an array literal.
// e.g.  [1, 2, 3]
// Token type: token.LBRACKET
type ArrayLiteral struct {
	Token    token.Token // the [ token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := make([]string, len(al.Elements))
	for i, el := range al.Elements {
		elements[i] = el.String()
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// HashLiteral represents a hash map literal.
// e.g.  {"name": "Alice", "age": 30}
// Keys and values are both Expressions.
// The colon separator corresponds to token.COLON in your token package.
// Token type: token.LBRACE
type HashLiteral struct {
	Token token.Token // the { token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
