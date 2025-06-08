package interpreter

import (
	"fmt"
	"github.com/ah-naf/borno/ast"
	"github.com/ah-naf/borno/environment"
	"github.com/ah-naf/borno/utils"
)

type Callable interface {
	Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error)
	Arity() int
}

type Function struct {
	Declaration   *ast.FunctionStmt
	Closure       *environment.Environment
	IsInitializer bool // New field
}

func NewFunction(declaration *ast.FunctionStmt, closure *environment.Environment) *Function {
	return &Function{Declaration: declaration, Closure: closure, IsInitializer: false}
}

func (f *Function) Bind(instance *BanglaInstance) Callable {
	methodEnv := environment.NewEnvironmentWithParent(f.Closure)
	methodEnv.Define("এইটা", instance) // "এইটা" is the Bangla keyword for 'this'
	return &Function{
		Declaration:   f.Declaration,
		Closure:       methodEnv,
		IsInitializer: f.IsInitializer,
	}
}

func (f *Function) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	execEnv := environment.NewEnvironmentWithParent(f.Closure)

	for idx, param := range f.Declaration.Params {
		execEnv.Define(param.Lexeme, arguments[idx])
	}

	for _, stmt := range f.Declaration.Body {
		// Ensure utils is imported for HadRuntimeError
		// Ensure fmt is imported for error messages
		if utils.HadRuntimeError { // Check before potentially complex eval
			return nil, fmt.Errorf("runtime error prior to statement execution in function %s", f.Declaration.Name.Lexeme)
		}
		_, signal := i.eval(stmt, execEnv, false)

		if signal.Type == ControlFlowReturn {
			if f.IsInitializer {
				if signal.Value == nil { // 'return;' from init
					thisVal, err := f.Closure.Get("এইটা") // 'this' is in f.Closure for bound methods
					if err != nil {
						return nil, fmt.Errorf("internal error: 'this' not found in initializer closure: %w", err)
					}
					return thisVal, nil
				}
				return signal.Value, nil // 'return someValue;' from init
			}
			return signal.Value, nil // Return from non-initializer
		}
		if signal.Type != ControlFlowNone {
			// Propagate error for unhandled break/continue
			return nil, fmt.Errorf("unhandled control flow signal (%d) escaped function %s", signal.Type, f.Declaration.Name.Lexeme)
		}
		if utils.HadRuntimeError { // Check after each statement's eval
			return nil, fmt.Errorf("runtime error during execution of function %s", f.Declaration.Name.Lexeme)
		}
	}

	if f.IsInitializer { // Implicit return from initializer
		thisVal, err := f.Closure.Get("এইটা") // 'this' is in f.Closure for bound methods
		if err != nil {
			return nil, fmt.Errorf("internal error: 'this' not found at end of initializer %s: %w", f.Declaration.Name.Lexeme, err)
		}
		return thisVal, nil
	}
	return nil, nil // Implicit return from non-initializer
}

func (f *Function) Arity() int {
	// Return the number of parameters the function takes.
	return len(f.Declaration.Params)
}

func (f *Function) String() string {
	return "<function " + f.Declaration.Name.Lexeme + ">"
}
