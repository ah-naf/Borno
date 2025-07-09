package interpreter

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type NativeClockFn struct{}

func (n NativeClockFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	return float64(time.Now().UnixMilli()) / 1000.0, nil
}

func (n NativeClockFn) Arity() int {
	return 0
}

func (n NativeClockFn) String() string {
	return "<native fn>"
}

// NativeInputFn defines the native `input` function for the interpreter.
type NativeInputFn struct{}

// Call executes the native `input` function.
func (n NativeInputFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	// Check if there's an optional prompt argument
	if len(arguments) > 1 {
		return nil, fmt.Errorf("input function accepts at most 1 argument")
	}

	// If a prompt argument is provided, print it
	if len(arguments) == 1 {
		var prompt string
		switch arg := arguments[0].(type) {
		case string:
			// Already a Go string
			prompt = arg
		case []rune:
			// Convert rune slice to string
			prompt = string(arg)
		default:
			return nil, fmt.Errorf("input function's argument must be a string or []rune")
		}
	
		fmt.Print(prompt)
	}

	// Read the input from the user
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %v", err)
	}

	// Trim the newline characters and return the input string
	return strings.TrimSpace(input), nil
}

func (n NativeInputFn) Arity() int {
	return -1 // Variable number of arguments: 0 or 1 (for prompt)
}

func (n NativeInputFn) String() string {
	return "<native fn input>"
}
