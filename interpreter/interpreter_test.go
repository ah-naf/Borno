package interpreter

import (
	"testing"

	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/lexer"
	"github.com/ah-naf/crafting-interpreter/parser"
	"github.com/ah-naf/crafting-interpreter/token"
	"github.com/ah-naf/crafting-interpreter/utils"
)

func TestEvalExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
		errorMsg string
	}{
		// Valid expressions
		{"Addition of numbers", "1 + 2", 3.0, ""},
		{"Subtraction of numbers", "5 - 2", 3.0, ""},
		{"Multiplication of numbers", "3 * 4", 12.0, ""},
		{"Division of numbers", "10 / 2", 5.0, ""},
		{"Comparison greater", "5 > 3", true, ""},
		{"Comparison less", "2 < 3", true, ""},
		{"Equality true", "4 == 4", true, ""},
		{"Equality false", "4 == 5", false, ""},
		{"Not equal true", "4 != 5", true, ""},
		{"Not equal false", "4 != 4", false, ""},
		{"Grouping and precedence", "(1 + 2) * 3", 9.0, ""},
		{"Unary minus", "-5", -5.0, ""},
		{"Unary bang true", "!true", false, ""},
		{"Unary bang false", "!false", true, ""},
		{"Unary bang number", "!0", false, ""},
		{"Nil equality", "nil == nil", true, ""},
		{"Addition of strings", "\"foo\" + \"bar\"", "foobar", ""},

		// Complex expressions with operator precedence
		{"Mixed precedence 1", "1 + 2 * 3", 7.0, ""},
		{"Mixed precedence 2", "(1 + 2) * 3", 9.0, ""},
		{"Complex precedence 1", "10 - 3 + 2 * 4 / 2", 11.0, ""},

		// Grouping
		{"Grouping expressions", "(1 + 2) * (3 + 4)", 21.0, ""},
		{"Nested grouping", "((1 + 2) * 3) + (4 * (5 - 2))", 21.0, ""},

		// Boolean expressions
		{"Boolean comparison", "true == false", false, ""},
		{"Boolean and number comparison", "true == 1", false, ""},

		// Nil-related expressions
		{"Nil equality", "nil == nil", true, ""},
		{"Nil addition", "nil + nil", nil, "Operands must be two numbers or two strings."},
		{"Nil in comparison", "nil > 1", nil, "Left operand must be a number."},

		// Complex arithmetic expressions
		{"Complex arithmetic 1", "((2 + 3) * 4 - 5) / 2", 7.5, ""},
		{"Complex arithmetic 2", "3 * (2 + (1 - 4) * (6 / 3))", -12.0, ""},

		// Error cases
		{"Division by zero", "10 / 0", nil, "Division by zero."},
		{"Invalid comparison with nil", "5 > nil", nil, "Right operand must be a number."},

		// String + number -> Should concatenate after converting number to string
		{"String and number concatenation", "\"Number: \" + 42", "Number: 42", ""},
		{"Number and string concatenation", "42 + \" is the answer\"", "42 is the answer", ""},

		// String + float
		{"String and float concatenation", "\"Pi is \" + 3.14", "Pi is 3.14", ""},

		// Number + string + number -> Should concatenate all
		{"Number + string + number", "123 + \" + \" + 456", "123 + 456", ""},

		// Invalid operations
		{"Invalid addition of string and boolean", "\"foo\" + true", nil, "Right operand must be a number or string."},
		{"Invalid addition of boolean and string", "true + \"foo\"", nil, "Left operand must be a number or string."},
		{"Invalid addition of string and nil", "\"foo\" + nil", nil, "Right operand must be a number or string."},
		{"Invalid addition of number and nil", "42 + nil", nil, "Right operand must be a number or string."},

		// Edge cases
		{"Nil addition with nil", "nil + nil", nil, "Operands must be two numbers or two strings."},
		{"Number and empty string", "42 + \"\"", "42", ""},
		{"Empty string and number", "\"\" + 42", "42", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset error flags before each test
			utils.HadError = false
			utils.HadRuntimeError = false

			// Lexical analysis
			scanner := lexer.NewScanner(tt.input)
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

			// Evaluation
			result := Eval(expr)

			if tt.errorMsg != "" {
				if !utils.HadRuntimeError {
					t.Errorf("Expected runtime error '%s', but got result %v", tt.errorMsg, result)
				}
			} else {
				if utils.HadRuntimeError {
					t.Errorf("Unexpected runtime error for input '%s'", tt.input)
				}
				if result != tt.expected {
					t.Errorf("For input '%s', expected %v, got %v", tt.input, tt.expected, result)
				}
			}
		})
	}
}

func TestEvalLiteral(t *testing.T) {
	tests := []struct {
		name     string
		literal  interface{}
		expected interface{}
	}{
		{"Number", 42.0, 42.0},
		{"String", "hello", "hello"},
		{"Boolean True", true, true},
		{"Boolean False", false, false},
		{"Nil", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Literal{Value: tt.literal}
			result := Eval(expr)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
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
		{"Negate Non-Number", token.MINUS, "hello", nil, "Operand must be a number."},
		{"Logical Not True", token.BANG, true, false, ""},
		{"Logical Not False", token.BANG, false, true, ""},
		{"Logical Not Nil", token.BANG, nil, true, ""},
		{"Logical Not Number", token.BANG, 42.0, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadRuntimeError = false

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

			result := Eval(expr)
			if tt.errorMsg != "" {
				if !utils.HadRuntimeError {
					t.Errorf("Expected runtime error '%s', but got result %v", tt.errorMsg, result)
				}
			} else {
				if utils.HadRuntimeError {
					t.Errorf("Unexpected runtime error")
				}
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
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
		{"Addition with Nil", nil, token.PLUS, 5.0, nil, "Operands must be two numbers or two strings."},
		{"Addition Nil + Nil", nil, token.PLUS, nil, nil, "Operands must be two numbers or two strings."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadRuntimeError = false

			leftExpr := &ast.Literal{Value: tt.left}
			rightExpr := &ast.Literal{Value: tt.right}

			operatorToken := token.Token{
				Type:    tt.operator,
				Lexeme:  tokenTypeToLexeme(tt.operator),
				Literal: nil,
				Line:    1,
			}

			expr := &ast.Binary{
				Left:     leftExpr,
				Operator: operatorToken,
				Right:    rightExpr,
			}

			result := Eval(expr)
			if tt.errorMsg != "" {
				if !utils.HadRuntimeError {
					t.Errorf("Expected runtime error '%s', but got result %v", tt.errorMsg, result)
				}
			} else {
				if utils.HadRuntimeError {
					t.Errorf("Unexpected runtime error")
				}
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestEvalGrouping(t *testing.T) {
	expr := &ast.Grouping{
		Expression: &ast.Literal{Value: 42.0},
	}
	result := Eval(expr)
	if result != 42.0 {
		t.Errorf("Expected 42.0, got %v", result)
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
