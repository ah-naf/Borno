package environment

import (
	"fmt"

	"github.com/ah-naf/crafting-interpreter/token"
	"github.com/ah-naf/crafting-interpreter/utils"
)

type Environment struct {
	Values map[string]interface{}
}

func NewEnvironment() *Environment {
	return &Environment{Values: make(map[string]interface{})}
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

	return nil, fmt.Errorf("undefined variable '%s'", name)
}

func (e *Environment) Assign(name token.Token, value interface{}) {
	if _, exists := e.Values[name.Lexeme]; exists {
		e.Values[name.Lexeme] = value
		return
	}
	utils.RuntimeError(name, "Undefined variable '"+name.Lexeme+"'.")
}
