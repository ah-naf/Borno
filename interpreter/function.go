package interpreter

import (
	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/environment"
)

type Callable interface {
	Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error)
	Arity() int
}

type Function struct {
	Declaration *ast.FunctionStmt
	Closure     *environment.Environment
}

func NewFunction(declaration *ast.FunctionStmt, closure *environment.Environment) *Function {
	return &Function{Declaration: declaration, Closure: closure}
}

func (f *Function) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	functionEnv := environment.NewEnvironmentWithParent(f.Closure)

	functionEnv.Define(f.Declaration.Name.Lexeme, f)

	for ind, param := range f.Declaration.Params {
		functionEnv.Define(param.Lexeme, arguments[ind])
	}

	for _, statment := range f.Declaration.Body {
		_, signal := i.eval(statment, functionEnv, false)
		if signal.Type == ControlFlowReturn {
			return signal.Value, nil
		}
		if signal.Type != ControlFlowNone {
			return nil, nil // You can later add support for return values.
		}
	}
	return nil, nil
}

func (f *Function) Arity() int {
	// Return the number of parameters the function takes.
	return len(f.Declaration.Params)
}

func (f *Function) String() string {
	return "<function " + f.Declaration.Name.Lexeme + ">"
}
