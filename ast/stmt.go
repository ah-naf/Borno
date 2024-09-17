package ast

import "fmt"

type Stmt interface {
	Expr
}

type ExpressionStatement struct {
	Expression Expr
}

type PrintStatement struct {
	Expression Expr
}

// String method for PrintStatement
func (p *PrintStatement) String() string {
	return fmt.Sprintf("(print %s)", p.Expression.String()) // Return string representation of print statement
}

// String method for ExpressionStatement
func (e *ExpressionStatement) String() string {
	return e.Expression.String() // Return string representation of the expression
}
