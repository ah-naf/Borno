package ast

import (
	"fmt"

	"github.com/ah-naf/borno/token"
)

type Stmt interface {
	Expr
}

type ExpressionStatement struct {
	Expression Expr
}

// String method for ExpressionStatement
func (e *ExpressionStatement) String() string {
	return e.Expression.String() // Return string representation of the expression
}

type PrintStatement struct {
	Expression Expr
}

// String method for PrintStatement
func (p *PrintStatement) String() string {
	return fmt.Sprintf("(print %s)", p.Expression.String()) // Return string representation of print statement
}

type VarStmt struct {
	Name        token.Token
	Initializer Expr
	// VarUsed		bool
	Line int
}

func (v *VarStmt) String() string {
	return fmt.Sprintf("var %s = %v", v.Name.Lexeme, v.Initializer)
}

type VarListStmt struct {
	Declarations []VarStmt
}

func (v *VarListStmt) String() string {
	output := ""
	for _, varStmt := range v.Declarations {
		output += fmt.Sprintf("var %s = %v\n", varStmt.Name.Lexeme, varStmt.Initializer)
	}
	return output
}

type AssignmentStmt struct {
	Name  token.Token
	Value Expr
	Line  int
}

func (a *AssignmentStmt) String() string {
	return fmt.Sprintf("(%s = %s)", a.Name.Lexeme, a.Value.String())
}

type BlockStmt struct {
	Block []Stmt
}

func (b *BlockStmt) String() string {
	val := fmt.Sprintf("{\n")
	for _, statement := range b.Block {
		val += fmt.Sprintf("%s\n", statement.String())
	}
	val += fmt.Sprint("}")
	return val
}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (i *IfStmt) String() string {
	val := fmt.Sprintf("if (%s)", i.Condition)
	val += i.ThenBranch.String()
	if i.ElseBranch != nil {
		val += "else "
		val += i.ElseBranch.String()

	}
	return val
}

type While struct {
	Condition Expr
	Body      Stmt
}

func (w *While) String() string {
	val := fmt.Sprintf("while (%s)", w.Condition)
	val += w.Body.String()
	return val
}

type ForStmt struct {
	Condition   Expr
	Increment   Expr
	Initializer Stmt
	Body        Stmt
}

func (f *ForStmt) String() string {
	initializerStr := ""
	if f.Initializer != nil {
		initializerStr = f.Initializer.String()
	}

	conditionStr := ""
	if f.Condition != nil {
		conditionStr = f.Condition.String()
	}

	incrementStr := ""
	if f.Increment != nil {
		incrementStr = f.Increment.String()
	}

	bodyStr := ""
	if f.Body != nil {
		bodyStr = f.Body.String()
	}

	return fmt.Sprintf("for (%v; %v; %v) %v", initializerStr, conditionStr, incrementStr, bodyStr)
}

type BreakStmt struct {
	Line int
}

func (b *BreakStmt) String() string {
	return "break"
}

type ContinueStmt struct {
	Line int
}

func (b *ContinueStmt) String() string {
	return "continue"
}

type FunctionStmt struct {
	Name   token.Token
	Params []token.Token
	Body   []Stmt
}

func (f *FunctionStmt) String() string {
	// Convert the list of parameters to a comma-separated string
	paramNames := ""
	for i, param := range f.Params {
		if i != 0 {
			paramNames += ", "
		}
		paramNames += param.Lexeme
	}

	// Convert the body statements to a string
	bodyStr := ""
	for _, stmt := range f.Body {
		bodyStr += stmt.String() + "\n"
	}

	// Return the function's string representation
	return fmt.Sprintf("fun %s(%s) {\n%s}", f.Name.Lexeme, paramNames, bodyStr)
}


type ArrayAssignment struct {
	Array Expr   // The array being assigned to
	Index Expr   // The index of the element being assigned to
	Value Expr   // The new value being assigned
	Line  int    // The line number of the assignment
}

func (a *ArrayAssignment) String() string {
	return fmt.Sprintf("(%s[%s] = %s)", a.Array, a.Index, a.Value)
}

// PropertyAssignment represents assigning a value to an object's property.
type PropertyAssignment struct {
	Object   Expr
	Property token.Token
	Value    Expr
	Line     int
}

func (p *PropertyAssignment) String() string {
	return fmt.Sprintf("%s.%s = %s", p.Object.String(), p.Property.Lexeme, p.Value.String())
}