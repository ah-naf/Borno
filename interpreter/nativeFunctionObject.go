package interpreter

import "fmt"

type NativeDeleteFn struct{}

func (n NativeDeleteFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	// Ensure we have exactly 2 arguments: the object and the key
	if len(arguments) != 2 {
		return nil, fmt.Errorf("delete function expects exactly 2 arguments (object and key)")
	}

	// Ensure the first argument is an object (map)
	object, ok := arguments[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("delete function only works on objects")
	}

	// Ensure the second argument is a string (key)
	var key string
	switch v := arguments[1].(type) {
	case string:
		key = v
	case []rune:
		key = string(v) // Convert []rune to string
	default:
		return nil, fmt.Errorf("delete function expects the second argument to be a string key")
	}

	// Remove the key if it exists
	if _, exists := object[key]; exists {
		delete(object, key)
	} else {
		return nil, fmt.Errorf("key '%s' not found in object", key)
	}

	return object, nil
}

func (n NativeDeleteFn) Arity() int {
	return 2 // Two arguments: object and key
}

func (n NativeDeleteFn) String() string {
	return "<native fn delete>"
}

type NativeKeysFn struct{}

func (n NativeKeysFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 1 {
		return nil, fmt.Errorf("keys function expects exactly 1 argument")
	}

	object, ok := arguments[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("keys function only works on objects")
	}

	keys := make([]interface{}, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}

	return keys, nil
}

func (n NativeKeysFn) Arity() int {
	return 1
}

func (n NativeKeysFn) String() string {
	return "<native fn keys>"
}

type NativeValuesFn struct{}

func (n NativeValuesFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 1 {
		return nil, fmt.Errorf("values function expects exactly 1 argument")
	}

	object, ok := arguments[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("values function only works on objects")
	}

	values := make([]interface{}, 0, len(object))
	for _, value := range object {
		values = append(values, value)
	}

	return values, nil
}

func (n NativeValuesFn) Arity() int {
	return 1
}

func (n NativeValuesFn) String() string {
	return "<native fn values>"
}
