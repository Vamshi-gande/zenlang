package ast

import "bytes"

// Node is the base interface every AST node must satisfy.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement is a node that performs an action but does not produce a usable value.
type Statement interface {
	Node
	statementNode()
}

// Expression is a node that produces a value.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST.
// The parser returns a single Program containing all top-level statements.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
