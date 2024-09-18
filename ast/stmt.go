package ast

import (
	"fmt"

	"github.com/ah-naf/crafting-interpreter/token"
)

type Stmt interface {
	Expr
}

type ExpressionStatement struct {
	Expression Expr
}

type PrintStatement struct {
	Expression Expr
}

type VarStmt struct {
	Name        token.Token
	Initializer Expr
	// VarUsed		bool
	Line int
}

// String method for PrintStatement
func (p *PrintStatement) String() string {
	return fmt.Sprintf("(print %s)", p.Expression.String()) // Return string representation of print statement
}

// String method for ExpressionStatement
func (e *ExpressionStatement) String() string {
	return e.Expression.String() // Return string representation of the expression
}

func (v *VarStmt) String() string {
	return fmt.Sprintf("var %s = %v", v.Name.Lexeme, v.Initializer)
}
