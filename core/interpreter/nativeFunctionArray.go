package interpreter

import "fmt"

type NativeLenFn struct{}

// Call executes the native `len` function
func (n NativeLenFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 1 {
		return nil, fmt.Errorf("len function expects exactly 1 argument")
	}

	// Check if the argument is a slice (array in our case)
	array, ok := arguments[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("len function only works on arrays")
	}

	// Return the length of the array
	return len(array), nil
}

func (n NativeLenFn) Arity() int {
	return 1 // The function expects one argument (the array)
}

func (n NativeLenFn) String() string {
	return "<native fn len>"
}

type NativeAppendFn struct{}

func (n NativeAppendFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) < 2 {
		return nil, fmt.Errorf("append function expects at least 2 arguments (array and element(s))")
	}

	// Ensure the first argument is an array
	array, ok := arguments[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("append function only works on arrays")
	}
	// Append all other arguments to the array
	array = append(array, arguments[1:]...)

	return array, nil
}

func (n NativeAppendFn) Arity() int {
	return -1 // Variable number of arguments (at least 2)
}

func (n NativeAppendFn) String() string {
	return "<native fn append>"
}

type NativeRemoveFn struct{}

func (n NativeRemoveFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 2 {
		return nil, fmt.Errorf("remove function expects exactly 2 arguments (array and index)")
	}

	// Ensure the first argument is an array
	array, ok := arguments[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("remove function only works on arrays")
	}

	// Ensure the second argument is an integer (index)
	index, err := toInt64(arguments[1])
	if err != nil {
		return nil, fmt.Errorf("array index must be an integer")
	}

	// Ensure the index is within bounds
	if index < 0 || int(index) >= len(array) {
		return nil, fmt.Errorf("array index out of bounds")
	}

	// Remove the element at the specified index
	array = append(array[:index], array[index+1:]...)

	return array, nil
}

func (n NativeRemoveFn) Arity() int {
	return 2 // Two arguments: array and index
}

func (n NativeRemoveFn) String() string {
	return "<native fn remove>"
}
