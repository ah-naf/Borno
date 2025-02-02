package interpreter

import (
	"bytes"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/lexer"
	"github.com/ah-naf/crafting-interpreter/parser"
	"github.com/ah-naf/crafting-interpreter/token"
	"github.com/ah-naf/crafting-interpreter/utils"
)

// CaptureStderr captures anything written to os.Stderr during the execution of the provided function.
func CaptureStderr(f func()) string {
	// Create a pipe to capture os.Stderr
	r, w, _ := os.Pipe()

	// Save the current os.Stderr so we can restore it later
	oldStderr := os.Stderr

	// Redirect os.Stderr to the pipe's writer
	os.Stderr = w

	// Run the provided function that might write to os.Stderr
	f()

	// Close the writer to stop capturing
	w.Close()

	// Restore the original os.Stderr
	os.Stderr = oldStderr

	// Read the captured output from the pipe
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Return the captured error output as a string
	return buf.String()
}

// Helper function to convert both int64 and float64 to float64 for comparison
func toFloat(val interface{}) interface{} {
	switch v := val.(type) {
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case float64:
		return v
	case string:
		ascii := utils.ConvertBanglaDigitsToASCII(v)
		num, _ := strconv.ParseFloat(ascii, 64)
		return num
	default:
		return val
	}
}

func TestEvalExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
		errorMsg string
	}{
		// New tests for bitwise operators
		{"Bitwise AND", "5 & 3;", int64(1), ""},
		{"Bitwise OR", "5 | 3;", int64(7), ""},
		{"Bitwise XOR", "5 ^ 3;", int64(6), ""},
		{"Left Shift", "2 << 1;", int64(4), ""},
		{"Right Shift", "8 >> 2;", int64(2), ""},
		{"Power", "3 ** 4;", int64(81), ""},

		// // Complex expressions involving bitwise and arithmetic
		{"Complex Bitwise and Arithmetic", "(5 & 3) + (8 >> 2) * 3 - (3 ** 2);", float64(1 + 6 - 9), ""},
		{"Complex Bitwise with Shift", "((5 | 2) << 1) + (8 ^ 3);", int64((7 << 1) + 11), ""},
		{"Complex Power and Shift", "2 ** (3 << 1);", int64(64), ""},
		{"Nested Grouping and Power", "((3 ** 2) + (8 >> 2)) * 2;", float64(11 * 2), ""},

		// // // Valid expressions
		{"Square root", "2 ** 0.5;", math.Pow(2, 0.5), ""},
		{"Modulo operator", "5 % 3;", 2.0, ""},
		{"Addition of numbers", "1 + 2;", 3.0, ""},
		{"Subtraction of numbers", "5 - 2;", 3.0, ""},
		{"Multiplication of numbers", "3 * 4;", 12.0, ""},
		{"Division of numbers", "10 / 2;", 5.0, ""},
		{"Comparison greater", "5 > 3;", true, ""},
		{"Comparison less", "2 < 3;", true, ""},
		{"Equality true", "4 == 4;", true, ""},
		{"Equality false", "4 == 5;", false, ""},
		{"Not equal true", "4 != 5;", true, ""},
		{"Not equal false", "4 != 4;", false, ""},
		{"Grouping and precedence", "(1 + 2) * 3;", 9.0, ""},
		{"Unary minus", "-5;", -5.0, ""},
		{"Unary bang true", "!সত্য;", false, ""},
		{"Unary bang false", "!মিথ্যা;", true, ""},
		{"Unary bang number", "!0;", true, ""},
		{"Nil equality", "nil == nil;", true, ""},
		{"Addition of strings", "\"foo\" + \"bar\";", "foobar", ""},

		// // // Complex expressions with operator precedence
		{"Mixed precedence 1", "1 + 2 * 3;", 7.0, ""},
		{"Mixed precedence 2", "(1 + 2) * 3;", 9.0, ""},
		{"Complex precedence 1", "10 - 3 + 2 * 4 / 2;", 11.0, ""},

		// // // Grouping
		{"Grouping expressions", "(1 + 2) * (3 + 4);", 21.0, ""},
		{"Nested grouping", "((1 + 2) * 3) + (4 * (5 - 2));", 21.0, ""},

		// // // Boolean expressions
		{"Boolean comparison", "সত্য == মিথ্যা;", false, ""},
		{"Boolean and number comparison", "সত্য == 1;", false, ""},

		// // // Nil-related expressions
		// {"Nil equality", "nil == nil;", true, ""},
		{"Nil addition", "nil + nil;", nil, "Operands must be numbers or strings."},
		{"Nil in comparison", "nil > 1;", nil, "Left operand must be a number."},

		// // Complex arithmetic expressions
		{"Complex arithmetic 1", "((2 + 3) * 4 - 5) / 2;", 7.5, ""},
		{"Complex arithmetic 2", "3 * (2 + (1 - 4) * (6 / 3));", -12.0, ""},

		// // Error cases
		{"Division by zero", "10 / 0;", nil, "Division by zero."},
		{"Invalid comparison with nil", "5 > nil;", nil, "Right operand must be a number."},

		// String + number -> Should concatenate after converting number to string
		{"String and number concatenation", "\"Number: \" + 42;", "Number: 42", ""},
		{"Number and string concatenation", "42 + \" is the answer\";", "42 is the answer", ""},

		// String + float
		{"String and float concatenation", "\"Pi is \" + 3.14;", "Pi is 3.14", ""},

		// Number + string + number -> Should concatenate all
		{"Number + string + number", "123 + \" + \" + 456;", "123 + 456", ""},

		// // Invalid operations
		{"Invalid addition of string and boolean", "\"foo\" + সত্য;", nil, "Right operand must be a string or number."},
		{"Invalid addition of boolean and string", "সত্য + \"foo\";", nil, "Operands must be numbers or strings."},
		{"Invalid addition of string and nil", "\"foo\" + nil;", nil, "Right operand must be a string or number."},
		{"Invalid addition of number and nil", "42 + nil;", nil, "Operands must be numbers or strings."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output interface{}

			// Reset error flags before each test
			utils.HadError = false
			utils.HadRuntimeError = false

			// Capture stderr during evaluation
			capturedErr := CaptureStderr(func() {
				// Lexical analysis
				scanner := lexer.NewScanner([]rune(tt.input))
				tokens := scanner.ScanTokens()

				// Check for lexical errors
				if utils.HadError {
					t.Fatalf("Scanner error for input '%s'", tt.input)
				}

				// Parsing
				parser := parser.NewParser(tokens)
				expr, err := parser.Parse()

				// Check for parsing errors
				if err != nil || utils.HadError {
					t.Fatalf("Parser error for input '%s'", tt.input)
				}

				interpreter := NewInterpreter()
				// Evaluation
				results := interpreter.Interpret(expr, false)
				// fmt.Println(results)
				if len(results) > 0 {
					output = results[0]
				}
			})

			// Handle runtime errors and comparison
			capturedErr = strings.Split(capturedErr, "\n")[0]

			if tt.errorMsg != "" {
				if !reflect.DeepEqual(capturedErr, tt.errorMsg) {
					t.Fatalf("Expected error %v, got %v", tt.errorMsg, capturedErr)
				}
			} else {
				// Ensure there is no runtime error when not expected
				if utils.HadRuntimeError {
					t.Fatalf("Unexpected runtime error for input '%s'", tt.input)
				}

				// Custom equality check for numbers
				if !reflect.DeepEqual(toFloat(output), toFloat(tt.expected)) {
					t.Fatalf("For input '%s', expected %v, got %v", tt.input, tt.expected, output)
				}
			}
		})
	}
}

func TestEvalUnary(t *testing.T) {
	tests := []struct {
		name     string
		operator token.TokenType
		operand  interface{}
		expected interface{}
		errorMsg string
	}{
		{"Negate Number", token.MINUS, 5.0, -5.0, ""},
		{"Negate Non-Number", token.MINUS, "hello", nil, `expected a number, got string "hello"`},
		{"Logical Not True", token.BANG, true, false, ""},
		{"Logical Not False", token.BANG, false, true, ""},
		{"Logical Not Nil", token.BANG, nil, true, ""},
		{"Logical Not Number", token.BANG, 42.0, false, ""},
		{"NOT operator", token.NOT, int64(1), -2, ""},
		{"NOT operator on string", token.NOT, "hello", nil, `expected an integer, got string "hello"`},
	}

	for _, tt := range tests {
		var output interface{}
		t.Run(tt.name, func(t *testing.T) {
			utils.HadRuntimeError = false

			// Capture stderr during evaluation
			capturedErr := CaptureStderr(func() {
				operatorToken := token.Token{
					Type:    tt.operator,
					Lexeme:  tokenTypeToLexeme(tt.operator),
					Literal: nil,
					Line:    1,
				}

				operandExpr := &ast.Literal{Value: tt.operand}
				expr := &ast.Unary{
					Operator: operatorToken,
					Right:    operandExpr,
				}

				interpreter := NewInterpreter()
				results := interpreter.Interpret([]ast.Stmt{expr}, false)
				if len(results) > 0 {
					output = results[0]
				}
			})

			capturedErr = strings.Split(capturedErr, "\n")[0]

			// Check for expected error messages
			if tt.errorMsg != "" {
				if capturedErr == "" || !utils.HadRuntimeError {
					t.Fatalf("Expected runtime error '%s', but got no error.", tt.errorMsg)
				} else if capturedErr != tt.errorMsg {
					t.Fatalf("Expected runtime error '%s', but got '%s'.", tt.errorMsg, capturedErr)
				}
			} else {
				// Ensure there is no runtime error when not expected
				if utils.HadRuntimeError {
					t.Fatalf("Unexpected runtime error for unary expression '%s'", tt.name)
				}
			}
			// fmt.Printf("%v, %v, %v\n", tt.expected, output, reflect.DeepEqual(toFloat(tt.expected), toFloat(output)))
			if !reflect.DeepEqual(toFloat(tt.expected), toFloat(output)) {
				t.Fatalf("Expected %v, got %v", tt.expected, output)
			}
		})
	}
}

func TestEvalBinary(t *testing.T) {
	tests := []struct {
		name     string
		left     interface{}
		operator token.TokenType
		right    interface{}
		expected interface{}
		errorMsg string
	}{
		{"Addition Numbers", 2.0, token.PLUS, 3.0, 5.0, ""},
		{"Addition Strings", "foo", token.PLUS, "bar", "foobar", ""},
		{"Addition Number and String", 2.0, token.PLUS, "bar", "2bar", ""},
		{"Subtraction", 5.0, token.MINUS, 2.0, 3.0, ""},
		{"Multiplication", 4.0, token.STAR, 2.5, 10.0, ""},
		{"Division", 10.0, token.SLASH, 2.0, 5.0, ""},
		{"Division by Zero", 10.0, token.SLASH, 0.0, nil, "Division by zero."},
		{"Greater Than", 5.0, token.GREATER, 3.0, true, ""},
		{"Less Than", 2.0, token.LESS, 3.0, true, ""},
		{"Equality True", 42.0, token.EQUAL_EQUAL, 42.0, true, ""},
		{"Equality False", 42.0, token.EQUAL_EQUAL, 43.0, false, ""},
		{"Inequality", "foo", token.BANG_EQUAL, "bar", true, ""},
		{"Comparison with Nil", nil, token.GREATER, 5.0, nil, "Left operand must be a number."},
		{"Addition with Nil", nil, token.PLUS, 5.0, nil, "Operands must be numbers or strings."},
		{"Addition Nil + Nil", nil, token.PLUS, nil, nil, "Operands must be numbers or strings."},
	}

	for _, tt := range tests {
		var output interface{}
		t.Run(tt.name, func(t *testing.T) {
			utils.HadRuntimeError = false

			// Capture stderr during evaluation
			capturedErr := CaptureStderr(func() {
				operatorToken := token.Token{
					Type:    tt.operator,
					Lexeme:  tokenTypeToLexeme(tt.operator),
					Literal: nil,
					Line:    1,
				}

				left := &ast.Literal{Value: tt.left}
				right := &ast.Literal{Value: tt.right}
				expr := &ast.Binary{
					Operator: operatorToken,
					Left:     left,
					Right:    right,
				}

				interpreter := NewInterpreter()
				results := interpreter.Interpret([]ast.Stmt{expr}, false)
				if len(results) > 0 {
					output = results[0]
				}
			})

			capturedErr = strings.Split(capturedErr, "\n")[0]

			// Check for expected error messages
			if tt.errorMsg != "" {
				if capturedErr == "" || !utils.HadRuntimeError {
					t.Fatalf("Expected runtime error '%s', but got no error.", tt.errorMsg)
				} else if capturedErr != tt.errorMsg {
					t.Fatalf("Expected runtime error '%s', but got '%s'.", tt.errorMsg, capturedErr)
				}
			} else {
				// Ensure there is no runtime error when not expected
				if utils.HadRuntimeError {
					t.Fatalf("Unexpected runtime error for unary expression '%s'", tt.name)
				}
			}
			// fmt.Printf("%v, %v, %v\n", tt.expected, output, reflect.DeepEqual(toFloat(tt.expected), toFloat(output)))
			if !reflect.DeepEqual(toFloat(tt.expected), toFloat(output)) {
				t.Fatalf("Expected %v, got %v", tt.expected, output)
			}
		})
	}
}

func tokenTypeToLexeme(tokenType token.TokenType) string {
	// Map token types to their lexemes for testing
	switch tokenType {
	case token.PLUS:
		return "+"
	case token.MINUS:
		return "-"
	case token.STAR:
		return "*"
	case token.SLASH:
		return "/"
	case token.BANG:
		return "!"
	case token.EQUAL_EQUAL:
		return "=="
	case token.BANG_EQUAL:
		return "!="
	case token.GREATER:
		return ">"
	case token.GREATER_EQUAL:
		return ">="
	case token.LESS:
		return "<"
	case token.LESS_EQUAL:
		return "<="
	default:
		return ""
	}
}
