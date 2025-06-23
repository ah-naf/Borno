package interpreter

// Class represents a user-defined class.
type Class struct {
	Name string
}

func NewClass(name string) *Class {
	return &Class{Name: name}
}

func (c *Class) String() string {
	return c.Name
}

// Arity returns the number of arguments required for instantiation.
func (c *Class) Arity() int {
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

// Call creates a new instance of the class.
func (c *Class) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	return NewInstance(c), nil
}


