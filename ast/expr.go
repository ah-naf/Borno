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
}

func (b *Binary) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator.Lexeme, b.Right.String())
}

// Grouping represents a grouped expression.
type Grouping struct {
	Expression Expr
}

func (g *Grouping) String() string {
	return fmt.Sprintf("(group %s)", g.Expression.String())
}

// Literal represents a literal value.
type Literal struct {
	Value interface{}
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

// Unary represents a unary expression.
type Unary struct {
	Operator token.Token
	Right    Expr
}

func (u *Unary) String() string {
	return fmt.Sprintf("(%s%s)", u.Operator.Lexeme, u.Right.String())
}
