package ast

import (
	"strconv"
	"strings"
)

// Expr is a node in an abstract syntax tree (AST).
type Expr interface {
	expr()
	AppendEquation(b []byte) []byte
}

type Const struct {
	Value float64
}

func (c *Const) expr() {}

func (c *Const) AppendEquation(b []byte) []byte {
	return strconv.AppendFloat(b, c.Value, 'g', -1, 64)
}

type Var struct {
	Ident string
}

func (v *Var) expr() {}

func (v *Var) AppendEquation(b []byte) []byte {
	return append(b, v.Ident...)
}

type Call struct {
	Fn   string
	Args []Expr
}

func (c *Call) expr() {}

func (c *Call) AppendEquation(b []byte) []byte {
	b = append(b, c.Fn...)
	b = append(b, '(')
	for i, arg := range c.Args {
		if i > 0 {
			b = append(b, ", "...)
		}
		b = arg.AppendEquation(b)
	}
	return append(b, ')')
}

type Subscript struct {
	Ident string
	Index []IExpr
}

func (s *Subscript) expr() {}

func (s *Subscript) AppendEquation(b []byte) []byte {
	b = append(b, s.Ident...)
	b = append(b, '[')
	for i, idx := range s.Index {
		if i > 0 {
			b = append(b, ", "...)
		}
		b = idx.AppendEquation(b)
	}
	return append(b, ']')
}

type UnaryOp struct {
	Op   string
	Expr Expr
}

func (u *UnaryOp) Valid() bool {
	const validBinOps = "+-"
	return strings.Contains(validBinOps, u.Op)
}

func (u *UnaryOp) expr() {}

func (u *UnaryOp) AppendEquation(b []byte) []byte {
	b = append(b, u.Op...)
	// Check if we need parentheses around the inner expression
	if _, ok := u.Expr.(*BinaryOp); ok {
		b = append(b, '(')
		b = u.Expr.AppendEquation(b)
		return append(b, ')')
	}
	return u.Expr.AppendEquation(b)
}

type BinaryOp struct {
	Op  string
	Lhs Expr
	Rhs Expr
}

func (b *BinaryOp) Valid() bool {
	const validBinOps = "+-*/"
	return strings.Contains(validBinOps, b.Op)
}

func (b *BinaryOp) expr() {}

// precedence returns the precedence level of the operator
func (b *BinaryOp) precedence() int {
	switch b.Op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

func (b *BinaryOp) AppendEquation(buf []byte) []byte {
	// Check if we need parentheses around the left operand
	if lhs, ok := b.Lhs.(*BinaryOp); ok && lhs.precedence() < b.precedence() {
		buf = append(buf, '(')
		buf = b.Lhs.AppendEquation(buf)
		buf = append(buf, ')')
	} else {
		buf = b.Lhs.AppendEquation(buf)
	}

	buf = append(buf, ' ')
	buf = append(buf, b.Op...)
	buf = append(buf, ' ')

	// Check if we need parentheses around the right operand
	// For right-associative operators (like subtraction and division), we need parens if precedence is equal or less
	if rhs, ok := b.Rhs.(*BinaryOp); ok {
		if rhs.precedence() < b.precedence() ||
			(rhs.precedence() == b.precedence() && (b.Op == "-" || b.Op == "/")) {
			buf = append(buf, '(')
			buf = b.Rhs.AppendEquation(buf)
			return append(buf, ')')
		}
	}
	return b.Rhs.AppendEquation(buf)
}

type If struct {
	Cond Expr
	Then Expr
	Else Expr
}

func (i *If) expr() {}

func (i *If) AppendEquation(b []byte) []byte {
	b = append(b, "if "...)
	b = i.Cond.AppendEquation(b)
	b = append(b, " then "...)
	b = i.Then.AppendEquation(b)
	b = append(b, " else "...)
	return i.Else.AppendEquation(b)
}

var (
	_ Expr = (*Const)(nil)
	_ Expr = (*Var)(nil)
	_ Expr = (*Call)(nil)
	_ Expr = (*Subscript)(nil)
	_ Expr = (*UnaryOp)(nil)
	_ Expr = (*BinaryOp)(nil)
	_ Expr = (*If)(nil)
)

// IExpr is an expression that appears inside "[" and "]" for subscripts.
type IExpr interface {
	indexExpr()
	AppendEquation(b []byte) []byte
}

type Wildcard struct{}

func (w *Wildcard) indexExpr() {}

func (w *Wildcard) AppendEquation(b []byte) []byte {
	return append(b, '*')
}

type StarRange struct {
	Ident string
}

func (s *StarRange) indexExpr() {}

func (s *StarRange) AppendEquation(b []byte) []byte {
	b = append(b, "*:"...)
	return append(b, s.Ident...)
}

type Range struct {
	Start Expr
	End   Expr
}

func (r *Range) indexExpr() {}

func (r *Range) AppendEquation(b []byte) []byte {
	b = r.Start.AppendEquation(b)
	b = append(b, ':')
	return r.End.AppendEquation(b)
}

type IndexExpr struct {
	Expr Expr
}

func (I *IndexExpr) indexExpr() {}

func (I *IndexExpr) AppendEquation(b []byte) []byte {
	return I.Expr.AppendEquation(b)
}

var (
	_ IExpr = (*Wildcard)(nil)
	_ IExpr = (*StarRange)(nil)
	_ IExpr = (*Range)(nil)
	_ IExpr = (*IndexExpr)(nil)
)
