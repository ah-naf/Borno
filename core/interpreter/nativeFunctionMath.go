package interpreter

import (
	"fmt"
	"math"
)

type NativeAbsFn struct{}

func (n NativeAbsFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 1 {
		return nil, fmt.Errorf("abs function expects exactly 1 argument")
	}

	number, err := toNumber(arguments[0])
	if err != nil {
		return nil, fmt.Errorf("argument must be a number")
	}

	return math.Abs(number), nil
}

func (n NativeAbsFn) Arity() int {
	return 1
}

func (n NativeAbsFn) String() string {
	return "<native fn abs>"
}

type NativeSqrtFn struct{}

func (n NativeSqrtFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 1 {
		return nil, fmt.Errorf("sqrt function expects exactly 1 argument")
	}

	number, err := toNumber(arguments[0])
	if err != nil {
		return nil, fmt.Errorf("argument must be a number")
	}

	return math.Sqrt(number), nil
}

func (n NativeSqrtFn) Arity() int {
	return 1
}

func (n NativeSqrtFn) String() string {
	return "<native fn sqrt>"
}

type NativePowFn struct{}

func (n NativePowFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 2 {
		return nil, fmt.Errorf("pow function expects exactly 2 arguments")
	}

	base, err := toNumber(arguments[0])
	if err != nil {
		return nil, fmt.Errorf("base must be a number")
	}

	exponent, err := toNumber(arguments[1])
	if err != nil {
		return nil, fmt.Errorf("exponent must be a number")
	}

	return math.Pow(base, exponent), nil
}

func (n NativePowFn) Arity() int {
	return 2
}

func (n NativePowFn) String() string {
	return "<native fn pow>"
}

type NativeSinFn struct{}

func (n NativeSinFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 1 {
		return nil, fmt.Errorf("sin function expects exactly 1 argument")
	}

	number, err := toNumber(arguments[0])
	if err != nil {
		return nil, fmt.Errorf("argument must be a number")
	}

	return math.Sin(number), nil
}

func (n NativeSinFn) Arity() int {
	return 1
}

func (n NativeSinFn) String() string {
	return "<native fn sin>"
}

type NativeCosFn struct{}

func (n NativeCosFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 1 {
		return nil, fmt.Errorf("cos function expects exactly 1 argument")
	}

	number, err := toNumber(arguments[0])
	if err != nil {
		return nil, fmt.Errorf("argument must be a number")
	}

	return math.Cos(number), nil
}

func (n NativeCosFn) Arity() int {
	return 1
}

func (n NativeCosFn) String() string {
	return "<native fn cos>"
}

type NativeTanFn struct{}

func (n NativeTanFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 1 {
		return nil, fmt.Errorf("tan function expects exactly 1 argument")
	}

	number, err := toNumber(arguments[0])
	if err != nil {
		return nil, fmt.Errorf("argument must be a number")
	}

	return math.Tan(number), nil
}

func (n NativeTanFn) Arity() int {
	return 1
}

func (n NativeTanFn) String() string {
	return "<native fn tan>"
}

// NativeMinFn defines the native `min` function for the interpreter.
type NativeMinFn struct{}

func (n NativeMinFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) == 0 {
		return nil, fmt.Errorf("min function expects at least 1 argument")
	}

	// Flatten arguments if the first argument is an array
	if array, ok := arguments[0].([]interface{}); ok && len(arguments) == 1 {
		arguments = array
	}

	if len(arguments) == 0 {
		return nil, fmt.Errorf("min function expects a non-empty array or list of arguments")
	}

	// Convert the first argument to a number
	minValue, err := toNumber(arguments[0])
	if err != nil {
		return nil, fmt.Errorf("all arguments must be numbers")
	}

	// Iterate over the remaining arguments
	for _, arg := range arguments[1:] {
		num, err := toNumber(arg)
		if err != nil {
			return nil, fmt.Errorf("all arguments must be numbers")
		}
		if num < minValue {
			minValue = num
		}
	}

	return minValue, nil
}

func (n NativeMinFn) Arity() int {
	return -1 // Variable number of arguments
}

func (n NativeMinFn) String() string {
	return "<native fn min>"
}

// NativeMaxFn defines the native `max` function for the interpreter.
type NativeMaxFn struct{}

func (n NativeMaxFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) == 0 {
		return nil, fmt.Errorf("max function expects at least 1 argument")
	}

	// Flatten arguments if the first argument is an array
	if array, ok := arguments[0].([]interface{}); ok && len(arguments) == 1 {
		arguments = array
	}

	if len(arguments) == 0 {
		return nil, fmt.Errorf("max function expects a non-empty array or list of arguments")
	}

	// Convert the first argument to a number
	maxValue, err := toNumber(arguments[0])
	if err != nil {
		return nil, fmt.Errorf("all arguments must be numbers")
	}

	// Iterate over the remaining arguments
	for _, arg := range arguments[1:] {
		num, err := toNumber(arg)
		if err != nil {
			return nil, fmt.Errorf("all arguments must be numbers")
		}
		if num > maxValue {
			maxValue = num
		}
	}

	return maxValue, nil
}

func (n NativeMaxFn) Arity() int {
	return -1 // Variable number of arguments
}

func (n NativeMaxFn) String() string {
	return "<native fn max>"
}

type NativeRoundFn struct{}

func (n NativeRoundFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	if len(arguments) != 1 {
		return nil, fmt.Errorf("round function expects exactly 1 argument")
	}

	number, err := toNumber(arguments[0])
	if err != nil {
		return nil, fmt.Errorf("argument must be a number")
	}

	return math.Round(number), nil
}

func (n NativeRoundFn) Arity() int {
	return 1
}

func (n NativeRoundFn) String() string {
	return "<native fn round>"
}
