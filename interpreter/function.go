package interpreter

import (
	"time"

	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/environment"
)

type Callable interface {
	Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error)
	Arity() int
}

type NativeClockFn struct{}

func (n NativeClockFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	return float64(time.Now().UnixMilli()) / 1000.0, nil
}

func (n NativeClockFn) Arity() int {
	return 0
}

func (n NativeClockFn) String() string {
	return "<native fn>"
}

type Function struct {
	Declaration *ast.FunctionStmt
}

func NewFunction(declaration *ast.FunctionStmt) *Function {
	return &Function{Declaration: declaration}
}

func (f *Function) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	functionEnv := environment.NewEnvironmentWithParent(i.globals)

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
