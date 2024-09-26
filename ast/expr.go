package ast

import (
	"fmt"

	"github.com/ah-naf/crafting-interpreter/token"
)

// Expr is the base interface for all expression types.
type Expr interface {
	String() string
}

// Binary represents a binary expression.
type Binary struct {
	Left     Expr
	Operator token.Token
	Right    Expr
	Line     int
}

func (b *Binary) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator.Lexeme, b.Right.String())
}

// Grouping represents a grouped expression.
type Grouping struct {
	Expression Expr
	Line       int
}

func (g *Grouping) String() string {
	return fmt.Sprintf("(group %s)", g.Expression.String())
}

// Literal represents a literal value.
type Literal struct {
	Value interface{}
	Line  int
}

func (l *Literal) String() string {
	if l.Value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", l.Value)
}

// Unary represents a unary expression.
type Unary struct {
	Operator token.Token
	Right    Expr
	Line     int
}

func (u *Unary) String() string {
	return fmt.Sprintf("(%s%s)", u.Operator.Lexeme, u.Right.String())
}

type Identifier struct {
	Name token.Token
	Line int
}

func (i *Identifier) String() string {
	return i.Name.Lexeme
}

type Logical struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func (l *Logical) String() string {
	return fmt.Sprintf("(%s %s %s)", l.Left.String(), l.Operator.Lexeme, l.Right.String())
}

// Call represents a function or method call expression.
type Call struct {
	Callee    Expr        // The expression that evaluates to the function (callee).
	Paren     token.Token // The opening parenthesis of the call (for error reporting).
	Arguments []Expr      // The list of arguments passed to the function.
}

func (c *Call) String() string {
	argStrings := ""
	for i, arg := range c.Arguments {
		if i != 0 {
			argStrings += ", "
		}
		argStrings += arg.String()
	}
	return fmt.Sprintf("%s(%s)", c.Callee.String(), argStrings)
}

type Return struct {
	Keyword token.Token
	Value   Expr
}

func (r *Return) String() string {
	return "return " + r.Value.String()
}

// ArrayLiteral represents an array literal in the source code.
type ArrayLiteral struct {
	Elements []Expr
	Line     int
}

func (a *ArrayLiteral) String() string {
	val := "["
	for i, e := range a.Elements {
		val += e.String()
		if i+1 != len(a.Elements) {
			val += ", "
		}
	}
	val += "]"
	return val
}

// ArrayAccess represents accessing an element from an array.
type ArrayAccess struct {
	Array Expr
	Index Expr
	Line  int
}

func (a *ArrayAccess) String() string {
	return fmt.Sprintf("%v[%v]", a.Array, a.Index)
}
