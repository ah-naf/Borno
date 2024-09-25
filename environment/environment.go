package environment

import (
	"fmt"

	"github.com/ah-naf/crafting-interpreter/token"
	"github.com/ah-naf/crafting-interpreter/utils"
)

type Environment struct {
	Values map[string]interface{}
	Parent *Environment
}

func NewEnvironment() *Environment {
	return &Environment{Values: make(map[string]interface{})}
}

func NewEnvironmentWithParent(parent *Environment) *Environment {
	return &Environment{Values: make(map[string]interface{}), Parent: parent}
}

// Define a new variable in environment
func (e *Environment) Define(name string, value interface{}) {
	e.Values[name] = value
}

// Get the value of a variable, checking parent scopes if necessary
func (e *Environment) Get(name string) (interface{}, error) {
	if value, exists := e.Values[name]; exists {
		return value, nil
	}

	if e.Parent != nil {
		return e.Parent.Get(name)
	}

	return nil, fmt.Errorf("undefined variable '%s'", name)
}

func (e *Environment) GetInCurrentScope(name string) (interface{}, error) {
    if value, exists := e.Values[name]; exists {
        return value, nil
    }
    return nil, fmt.Errorf("undefined variable '%s'", name)
}


func (e *Environment) Assign(name token.Token, value interface{}) {
	if _, exists := e.Values[name.Lexeme]; exists {
		e.Values[name.Lexeme] = value
		return
	}

	if e.Parent != nil {
		e.Parent.Assign(name, value)
		return
	}

	utils.RuntimeError(name, "Undefined variable '"+name.Lexeme+"'.")
}
