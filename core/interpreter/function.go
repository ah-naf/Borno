package interpreter

import (
	"github.com/ah-naf/borno/ast"
	"github.com/ah-naf/borno/environment"
	"github.com/ah-naf/borno/token"
	"github.com/ah-naf/borno/utils"
)

type Callable interface {
	Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error)
	Arity() int
}

type Function struct {
	Declaration   *ast.FunctionStmt
	Closure       *environment.Environment
	isInitializer bool
}

func NewFunction(declaration *ast.FunctionStmt, closure *environment.Environment) *Function {
	return &Function{Declaration: declaration, Closure: closure}
}

func (f *Function) Bind(instance *Instance) *Function {
	env := environment.NewEnvironmentWithParent(f.Closure)
	env.Define("this", instance)
	return &Function{Declaration: f.Declaration, Closure: env, isInitializer: f.isInitializer}
}

func (f *Function) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	functionEnv := environment.NewEnvironmentWithParent(f.Closure)

	functionEnv.Define(f.Declaration.Name.Lexeme, f)

	for ind, param := range f.Declaration.Params {
		functionEnv.Define(param.Lexeme, arguments[ind])
	}

	var ret interface{}
	var retSignal *ControlFlowSignal
	for _, statment := range f.Declaration.Body {
		_, signal := i.eval(statment, functionEnv, false)
		if signal.Type == ControlFlowReturn {
			ret = signal.Value
			retSignal = signal
			break
		}
		if signal.Type != ControlFlowNone {
			return nil, nil // You can later add support for return values.
		}
	}
	if f.isInitializer {
		if retSignal != nil && retSignal.Value != nil {
			tok := token.Token{Type: token.RETURN, Lexeme: "return", Line: retSignal.LineNumber}
			utils.GlobalErrorToken(tok, "Can't return a value from an initializer.")
			return nil, nil
		}
		val, _ := functionEnv.Get("this")
		return val, nil
	}
	return ret, nil
}

func (f *Function) Arity() int {
	// Return the number of parameters the function takes.
	return len(f.Declaration.Params)
}

func (f *Function) String() string {
	return "<function " + f.Declaration.Name.Lexeme + ">"
}
