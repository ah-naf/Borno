package interpreter

import (
	"fmt"
	"os" // For error exits, similar to Lox example. Consider if error handling should be different.

	"github.com/ah-naf/borno/environment" // For Environment in Bind method of UserFunction (anticipating next step)
	"github.com/ah-naf/borno/token"
	// "github.com/ah-naf/borno/ast" // Not directly needed here, but UserFunction might refer to it.
)

// Forward declaration for Interpreter, will be properly imported or passed if needed by methods.
// type Interpreter struct { ... } // Not defining here, just acknowledging it's used by Call.

// BanglaClass represents the runtime representation of a class.
type BanglaClass struct {
	Name       string
	Superclass *BanglaClass
	Methods    map[string]*Function // Ensure this is *Function
}

// String returns the class name when printed.
func (c *BanglaClass) String() string {
	return c.Name
}

// Arity returns the number of arguments expected by the class's initializer.
// If there's no 'init' method, it's 0.
// func (c *BanglaClass) Arity(interpreter *Interpreter) int {
func (c *BanglaClass) Arity() int { // Changed signature
	if initializer, ok := c.Methods["init"]; ok { // Methods map should store *Function
		return initializer.Arity()
	}
    if c.Superclass != nil { // Check superclass for init if not in current class
        // This part of arity check through superclass might be tricky if super.init has different arity
        // Lox doesn't explicitly show arity check propagating like this for Call,
        // but rather findInitializer propagates. Let's stick to current class init arity.
        // For Call, initializer is found recursively.
    }
	return 0
}

// Call creates an instance of the class and calls its initializer.
func (c *BanglaClass) Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error) {
	instance := &BanglaInstance{
		Klass:  c,
		Fields: make(map[string]interface{}),
	}

	// Find the initializer ("init" method)
	if initializerFunc, ok := c.FindMethod("init"); ok { // initializerFunc is *Function
		boundInit := initializerFunc.Bind(instance) // boundInit is Callable
		// The result of init's Call is often secondary to the instance itself.
		// However, if init errors, we should propagate that.
		_, callErr := boundInit.Call(interpreter, arguments)
		if callErr != nil {
			// Lox often returns the instance even if init fails, but also reports the error.
			// For robustness, let's return the instance and the error.
			return instance, callErr
		}
	} else if len(arguments) > 0 {
		// No 'init' method, but arguments were provided.
		// This should ideally be caught by an arity check before Call is invoked.
		// If LoxClass.Arity() returned 0, and arguments are passed, VisitCallExpr should error.
		// If we reach here, it implies a mismatch or a design choice to allow it.
		// For now, let's assume Arity check handles this. If not, an error could be raised here.
		// return nil, fmt.Errorf("class %s does not have an init method but received %d arguments", c.Name, len(arguments))
	}
	return instance, nil // Return instance, and nil error if init was successful or no init
}

// FindMethod looks for a method in this class or any superclass.
func (c *BanglaClass) FindMethod(name string) (*Function, bool) { // Ensure this returns *Function
	if method, ok := c.Methods[name]; ok {
		return method, true
	}
	if c.Superclass != nil {
		return c.Superclass.FindMethod(name)
	}
	return nil, false
}

// BanglaInstance represents an instance of a BanglaClass.
type BanglaInstance struct {
	Klass  *BanglaClass
	Fields map[string]interface{}
}

// String returns a string representation of an instance.
func (i *BanglaInstance) String() string {
	return fmt.Sprintf("%s instance", i.Klass.Name)
}

// Get retrieves a property (field or method) from the instance.
// Methods are bound to the instance upon retrieval.
func (inst *BanglaInstance) Get(name token.Token, interpreter *Interpreter) (interface{}, error) {
	// Look for field first.
	if value, ok := inst.Fields[name.Lexeme]; ok {
		return value, nil
	}

	// If no field, look for a method.
	if method, ok := inst.Klass.FindMethod(name.Lexeme); ok {
		return method.Bind(inst), nil // Bind will be added to UserFunction
	}

	// No field and no method found.
	// Using fmt.Fprintf to os.Stderr and os.Exit is how Lox handles runtime errors.
	// Consider returning an error value instead for better testability / Go style.
	// For now, replicate Lox behavior.
	// runtimeErr := NewRuntimeError(name, fmt.Sprintf("Undefined property '%s'.", name.Lexeme))
	// interpreter.runtimeError(runtimeErr) // Assuming interpreter has a runtimeError method
	// return nil, runtimeErr
	// The issue description's LoxInstance.Get uses os.Exit.
	fmt.Fprintf(os.Stderr, "[line %d] Error: Undefined property '%s'.\n", name.Line, name.Lexeme)
	os.Exit(70) // Exit code for runtime error
	return nil, nil // Unreachable after os.Exit
}

// Set assigns a value to a field on the instance.
func (inst *BanglaInstance) Set(name token.Token, value interface{}) {
	inst.Fields[name.Lexeme] = value
}

// --- Helper or parts related to Callable interface (might be needed for UserFunction.Bind) ---
// UserFunction.Bind will return a Callable, so the result of Bind needs a Call method.
// This implies that the result of Bind (often another UserFunction or a special bound method type)
// must also implement the Callable interface.
// For now, UserFunction itself is assumed to be Callable.
// The Bind method on UserFunction will create a new UserFunction with the 'this' environment.

// Ensure `UserFunction` (from `interpreter/function.go`) has or will have:
// - An `Arity(*Interpreter) int` method. (It seems it has `Arity() int` from Lox, adapt if needed)
// - A `Call(*Interpreter, []interface{}) (interface{}, error)` method.
// - A `Bind(*BanglaInstance) *UserFunction` method (to be added in next step).

// Make sure necessary packages like `fmt`, `os`, `token` are imported.
// `environment` is imported in anticipation of its use in `UserFunction.Bind`.
