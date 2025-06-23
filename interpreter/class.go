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
