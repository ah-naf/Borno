package interpreter

import (
	"fmt"

	"github.com/ah-naf/borno/token"
)

// Class represents a user-defined class.
type Class struct {
	Name    string
	Methods map[string]*Function
}

func NewClass(name string, methods map[string]*Function) *Class {
	return &Class{Name: name, Methods: methods}
}

func (c *Class) String() string {
	return c.Name
}

// Arity returns the number of arguments required for instantiation.
func (c *Class) Arity() int {
	if initializer, ok := c.Methods["init"]; ok {
		return initializer.Arity()
	}
	return 0
}

type Instance struct {
	Class  *Class
	Fields map[string]interface{}
}

// NewInstance creates a new instance of the given class.
func NewInstance(class *Class) *Instance {
	return &Instance{Class: class, Fields: make(map[string]interface{})}
}

// String returns the textual representation of the instance.
func (i *Instance) String() string {
	if i.Class == nil {
		return "instance"
	}
	return i.Class.Name + " instance"
}

func (i *Instance) Get(name token.Token) (interface{}, error) {
	if value, ok := i.Fields[name.Lexeme]; ok {
		return value, nil
	}
	if method, ok := i.Class.Methods[name.Lexeme]; ok {
		return method.Bind(i), nil
	}
	return nil, fmt.Errorf("Undefined property '%s'.", name.Lexeme)
}

// Set assigns a value to a property by name.
func (i *Instance) Set(name token.Token, value interface{}) {
	i.Fields[name.Lexeme] = value
}

// Call creates a new instance of the class.
func (c *Class) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	instance := NewInstance(c)
	if initializer, ok := c.Methods["init"]; ok {
		_, _ = initializer.Bind(instance).Call(i, arguments)
	}
	return instance, nil
}
