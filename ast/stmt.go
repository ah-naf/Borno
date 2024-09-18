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

type VarListStmt struct {
	Declarations []VarStmt
}

type AssignmentStmt struct {
	Name  token.Token
	Value Expr
	Line  int
}

type BlockStmt struct {
	Block []Stmt
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

func (v *VarListStmt) String() string {
	output := ""
	for _, varStmt := range v.Declarations {
		output += fmt.Sprintf("var %s = %v\n", varStmt.Name.Lexeme, varStmt.Initializer)
	}
	return output
}

func (a *AssignmentStmt) String() string {
	return fmt.Sprintf("(%s = %s)", a.Name.Lexeme, a.Value.String())
}

func (b *BlockStmt) String() string {
	val := fmt.Sprintf("{\n")
	for _, statement := range b.Block {
		val += fmt.Sprintf("%s\n", statement.String())
	}
	val += fmt.Sprint("}")
	return val
}
