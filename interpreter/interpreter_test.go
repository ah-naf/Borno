package interpreter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ah-naf/borno/ast"
	"github.com/ah-naf/borno/lexer"
	"github.com/ah-naf/borno/parser"
	"github.com/ah-naf/borno/token"
	"github.com/ah-naf/borno/utils"
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

func TestClassInstantiation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple instance",
			input:    "class Spaceship {}; ধরি falcon = Spaceship(); falcon;",
			expected: "Spaceship instance",
		},
		{
			name:     "Multiple instances",
			input:    "class Robot {}; ধরি r1 = Robot(); ধরি r2 = Robot(); r1; r2;",
			expected: "Robot instance",
		},
		{
			name:     "Instance from function",
			input:    "class Wizard {}; class Dragon {}; ফাংশন createCharacters() { ধরি merlin = Wizard(); ধরি smaug = Dragon(); ফেরত merlin; } ধরি mainCharacter = createCharacters(); mainCharacter;",
			expected: "Wizard instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadError = false
			utils.HadRuntimeError = false

			scanner := lexer.NewScanner([]rune(tt.input))
			tokens := scanner.ScanTokens()

			parser := parser.NewParser(tokens)
			stmts, err := parser.Parse()
			if err != nil || utils.HadError {
				t.Fatalf("Parser error: %v", err)
			}

			interp := NewInterpreter()
			results := interp.Interpret(stmts, false)

			if utils.HadRuntimeError {
				t.Fatalf("Runtime error while executing %s", tt.name)
			}

			if len(results) == 0 {
				t.Fatalf("No result returned")
			}

			result := results[len(results)-1]
			str := stringify(result)
			if str != tt.expected {
				t.Fatalf("Expected %s, got %s", tt.expected, str)
			}
		})
	}
}

func TestInstanceProperties(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interface{}
	}{
		{
			name: "Basic property get/set",
			input: `
class Spaceship {};
var falcon = Spaceship();

falcon.name = "Millennium Falcon";
falcon.speed = 75.5;

print falcon.name;
print falcon.speed;
			`,
			expected: []interface{}{"Millennium Falcon", 75.5},
		},
		{
			name: "Multiple instances properties",
			input: `
class Robot {};
var r2d2 = Robot();

r2d2.model = "Astromech";
r2d2.operational = false;

if (r2d2.operational) {
  print r2d2.model;
  r2d2.mission = "Navigate hyperspace";
  print r2d2.mission;
}
			`,
			expected: []interface{}{},
		},
		{
			name: "Property manipulation in function",
			input: `
// Multiple instances with properties
class Superhero {};
var batman = Superhero();
var superman = Superhero();

batman.name = "Batman";
batman.called = 18;

superman.name = "Superman";
superman.called = 66;

print "Times " + superman.name + " was called: ";
print superman.called;
print "Times " + batman.name + " was called: ";
print batman.called;
			`,
			expected: []interface{}{"Times Superman was called: ", int64(66), "Times Batman was called: ", int64(18)},
		},
		{
			name: "Another property manipulation in function",
			input: `
// Property manipulation in functions
class Wizard {};
var gandalf = Wizard();

gandalf.color = "Grey";
gandalf.power = nil;
print gandalf.color;

fun promote(wizard) {
  wizard.color = "White";
  if (true) {
    wizard.power = 100;
  } else {
    wizard.power = 0;
  }
}

promote(gandalf);
print gandalf.color;
print gandalf.power;
			`,
			expected: []interface{}{"Grey", "White", int64(100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadError = false
			utils.HadRuntimeError = false

			// 1) lex & parse
			scanner := lexer.NewScanner([]rune(tt.input))
			tokens := scanner.ScanTokens()

			parser := parser.NewParser(tokens)
			stmts, err := parser.Parse()
			if err != nil || utils.HadError {
				t.Fatalf("Parser error: %v", err)
			}

			// 2) capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			// 3) run the interpreter
			interp := NewInterpreter()
			_ = interp.Interpret(stmts, false)

			// 4) restore stdout and read captured output
			w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			if utils.HadRuntimeError {
				t.Fatalf("Runtime error while executing %s", tt.name)
			}

			// 5) split into lines and compare
			output := strings.TrimSpace(buf.String())
			lines := []string{}
			scannerOut := bufio.NewScanner(strings.NewReader(output))
			for scannerOut.Scan() {
				lines = append(lines, scannerOut.Text())
			}
			if err := scannerOut.Err(); err != nil {
				t.Fatalf("Error reading output: %v", err)
			}

			if len(lines) != len(tt.expected) {
				t.Fatalf("Expected %d lines of output, got %d: %v",
					len(tt.expected), len(lines), lines)
			}

			for i, exp := range tt.expected {
				// convert expected to its printed form
				expStr := fmt.Sprint(exp)
				if lines[i] != expStr {
					t.Errorf("line %d: expected %q, got %q", i+1, expStr, lines[i])
				}
			}
		})
	}
}

func TestInstanceMethods(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interface{}
	}{
		{
			name: "Return class name",
			input: `
class Foo {
  returnSelf() {
    return Foo;
  }
};

print Foo().returnSelf();
                        `,
			expected: []interface{}{"Foo"},
		},
		{
			name: "Method variables",
			input: `
class Wizard {
  castSpell(spell) {
    print "Casting a magical spell: " + spell;
  }
};

class Dragon {
  breatheFire(fire, intensity) {
    print "Breathing " + fire + " with intensity: " + intensity;
  }
};

var merlin = Wizard();
var smaug = Dragon();

if (true) {
  var action = merlin.castSpell;
  action("Fireball");
} else {
  var action = smaug.breatheFire;
  action("Fire", "100");
}
                        `,
			expected: []interface{}{"Casting a magical spell: Fireball"},
		},
		{
			name: "Methods in functions",
			input: `
class Superhero {
  useSpecialPower(hero) {
    print "Using power: " + hero.specialPower;
  }

  hasSpecialPower(hero) {
    return hero.specialPower;
  }

  giveSpecialPower(hero, power) {
    hero.specialPower = power;
  }
};

fun performHeroics(hero, superheroClass) {
  if (superheroClass.hasSpecialPower(hero)) {
    superheroClass.useSpecialPower(hero);
  } else {
    print "No special power available";
  }
}

var superman = Superhero();
var heroClass = Superhero();

if (true) {
  heroClass.giveSpecialPower(superman, "Flight");
} else {
  heroClass.giveSpecialPower(superman, "Strength");
}

performHeroics(superman, heroClass);
                        `,
			expected: []interface{}{"Using power: Flight"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadError = false
			utils.HadRuntimeError = false

			scanner := lexer.NewScanner([]rune(tt.input))
			tokens := scanner.ScanTokens()

			parser := parser.NewParser(tokens)
			stmts, err := parser.Parse()
			if err != nil || utils.HadError {
				t.Fatalf("Parser error: %v", err)
			}

			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			interp := NewInterpreter()
			_ = interp.Interpret(stmts, false)

			w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			if utils.HadRuntimeError {
				t.Fatalf("Runtime error while executing %s", tt.name)
			}

			output := strings.TrimSpace(buf.String())
			lines := []string{}
			scannerOut := bufio.NewScanner(strings.NewReader(output))
			for scannerOut.Scan() {
				lines = append(lines, scannerOut.Text())
			}
			if err := scannerOut.Err(); err != nil {
				t.Fatalf("Error reading output: %v", err)
			}

			if len(lines) != len(tt.expected) {
				t.Fatalf("Expected %d lines of output, got %d: %v", len(tt.expected), len(lines), lines)
			}

			for i, exp := range tt.expected {
				expStr := fmt.Sprint(exp)
				if lines[i] != expStr {
					t.Errorf("line %d: expected %q, got %q", i+1, expStr, lines[i])
				}
			}
		})
	}
}

func TestThisKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interface{}
	}{
		{
			name: "Basic this",
			input: `
class Spaceship {
  identify() {
    print this;
  }
};

Spaceship().identify();
`,
			expected: []interface{}{"Spaceship instance"},
		},
		{
			name: "Instance property access",
			input: `
class Calculator {
  add(a, b) {
    return a + b + this.memory;
  }
};

var calc = Calculator();
calc.memory = 82;
print calc.add(92, 1);
`,
			expected: []interface{}{int64(175)},
		},
		{
			name: "Method binding",
			input: `
class Animal {
  makeSound() {
    print this.sound;
  }
  identify() {
    print this.species;
  }
};

var dog = Animal();
dog.sound = "Woof";
dog.species = "Dog";

var cat = Animal();
cat.sound = "Meow";
cat.species = "Cat";

cat.makeSound = dog.makeSound;
dog.identify = cat.identify;

cat.makeSound();
dog.identify();
`,
			expected: []interface{}{"Woof", "Cat"},
		},
		{
			name: "Nested function this",
			input: `
class Wizard {
  getSpellCaster() {
    fun castSpell() {
      print this;
      print "Casting spell as " + this.name;
    }

    return castSpell;
  }
};

var wizard = Wizard();
wizard.name = "Merlin";
wizard.getSpellCaster()();
`,
			expected: []interface{}{"Wizard instance", "Casting spell as Merlin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadError = false
			utils.HadRuntimeError = false

			scanner := lexer.NewScanner([]rune(tt.input))
			tokens := scanner.ScanTokens()

			parser := parser.NewParser(tokens)
			stmts, err := parser.Parse()
			if err != nil || utils.HadError {
				t.Fatalf("Parser error: %v", err)
			}

			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			interp := NewInterpreter()
			_ = interp.Interpret(stmts, false)

			w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			if utils.HadRuntimeError {
				t.Fatalf("Runtime error while executing %s", tt.name)
			}

			output := strings.TrimSpace(buf.String())
			lines := []string{}
			scannerOut := bufio.NewScanner(strings.NewReader(output))
			for scannerOut.Scan() {
				lines = append(lines, scannerOut.Text())
			}
			if err := scannerOut.Err(); err != nil {
				t.Fatalf("Error reading output: %v", err)
			}

			if len(lines) != len(tt.expected) {
				t.Fatalf("Expected %d lines of output, got %d: %v", len(tt.expected), len(lines), lines)
			}

			for i, exp := range tt.expected {
				expStr := fmt.Sprint(exp)
				if lines[i] != expStr {
					t.Errorf("line %d: expected %q, got %q", i+1, expStr, lines[i])
				}
			}
		})
	}
}

func TestInvalidThisUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		errorMsg string
	}{
		{
			name: "Top level this",
			input: `
print this;
`,
			errorMsg: "Can't use 'this' outside of a class.",
		},
		{
			name: "This in function",
			input: `
fun notAMethod() {
  print this;
}
notAMethod();
`,
			errorMsg: "Can't use 'this' outside of a class.",
		},
		{
			name: "Call this as function",
			input: `
class Person {
  sayName() {
    print this();
  }
};
Person().sayName();
`,
			errorMsg: "Can only call functions and classes.",
		},
		{
			name: "Invalid property on this",
			input: `
class Confused {
  method() {
    fun inner(instance) {
      var feeling = "confused";
      print this.feeling;
    }
    return inner;
  }
};

var instance = Confused();
var m = instance.method();
m(instance);
`,
			errorMsg: "Undefined property 'feeling'.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadError = false
			utils.HadRuntimeError = false

			scanner := lexer.NewScanner([]rune(tt.input))
			tokens := scanner.ScanTokens()

			parser := parser.NewParser(tokens)
			stmts, err := parser.Parse()
			if err != nil || utils.HadError {
				t.Fatalf("Parser error: %v", err)
			}

			capturedErr := CaptureStderr(func() {
				interp := NewInterpreter()
				_ = interp.Interpret(stmts, false)
			})

			capturedErr = strings.Split(capturedErr, "\n")[0]

			if !utils.HadRuntimeError {
				t.Fatalf("Expected runtime error for %s", tt.name)
			}
			if capturedErr != tt.errorMsg {
				t.Fatalf("Expected error %q, got %q", tt.errorMsg, capturedErr)
			}
		})
	}
}

func TestConstructors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "Default properties",
			input: `
class Default {
  init() {
    this.x = "bar";
    this.y = 91;
  }
};
print Default().x;
print Default().y;
`,
			expected: []string{"bar", "91"},
		},
		{
			name: "Constructor parameters",
			input: `
class Robot {
  init(model, function) {
    this.model = model;
    this.function = function;
  }
};
print Robot("R2-D2", "Astromech").model;
`,
			expected: []string{"R2-D2"},
		},
		{
			name: "Initializer chaining",
			input: `
class Counter {
  init(startValue) {
    if (startValue < 0) {
      print "startValue can't be negative";
      this.count = 0;
    } else {
      this.count = startValue;
    }
  }
};

var instance = Counter(-52);
print instance.count;
print instance.init(52).count;
`,
			expected: []string{"startValue can't be negative", "0", "52"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadError = false
			utils.HadRuntimeError = false

			scanner := lexer.NewScanner([]rune(tt.input))
			tokens := scanner.ScanTokens()

			parser := parser.NewParser(tokens)
			stmts, err := parser.Parse()
			if err != nil || utils.HadError {
				t.Fatalf("Parser error: %v", err)
			}

			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			interp := NewInterpreter()
			_ = interp.Interpret(stmts, false)

			w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			if utils.HadRuntimeError {
				t.Fatalf("Runtime error while executing %s", tt.name)
			}

			output := strings.TrimSpace(buf.String())
			lines := []string{}
			scannerOut := bufio.NewScanner(strings.NewReader(output))
			for scannerOut.Scan() {
				lines = append(lines, scannerOut.Text())
			}
			if err := scannerOut.Err(); err != nil {
				t.Fatalf("Error reading output: %v", err)
			}

			if len(lines) != len(tt.expected) {
				t.Fatalf("Expected %d lines of output, got %d: %v", len(tt.expected), len(lines), lines)
			}

			for i, exp := range tt.expected {
				if lines[i] != exp {
					t.Errorf("line %d: expected %q, got %q", i+1, exp, lines[i])
				}
			}
		})
	}
}

func TestReturnInConstructors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		errorMsg string
	}{
		{
			name: "Return without value",
			input: `
class Person {
  init() {
    print "world";
    return;
  }
};

Person();
`,
			expected: []string{"world"},
		},
		{
			name: "Return this value",
			input: `
class ThingDefault {
  init() {
    this.x = "foo";
    this.y = 42;
    return this;
  }
};
var out = ThingDefault();
print out;
`,
			errorMsg: "[line 6] Error at 'return': Can't return a value from an initializer.",
		},
		{
			name: "Return string value",
			input: `
class Foo {
  init() {
    return "something else";
  }
};

Foo();
`,
			errorMsg: "[line 4] Error at 'return': Can't return a value from an initializer.",
		},
		{
			name: "Return call",
			input: `
class Foo {
  init() {
    return this.callback();
  }

  callback() {
    return "callback";
  }
};

Foo();
`,
			errorMsg: "[line 4] Error at 'return': Can't return a value from an initializer.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadError = false
			utils.HadRuntimeError = false

			scanner := lexer.NewScanner([]rune(tt.input))
			tokens := scanner.ScanTokens()

			parser := parser.NewParser(tokens)
			stmts, err := parser.Parse()
			if err != nil || utils.HadError {
				t.Fatalf("Parser error: %v", err)
			}

			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			capturedErr := CaptureStderr(func() {
				interp := NewInterpreter()
				_ = interp.Interpret(stmts, false)
			})

			w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			capturedErr = strings.Split(capturedErr, "\n")[0]

			if tt.errorMsg != "" {
				if !utils.HadError {
					t.Fatalf("Expected compile error for %s", tt.name)
				}
				if capturedErr != tt.errorMsg {
					t.Fatalf("Expected error %q, got %q", tt.errorMsg, capturedErr)
				}
				return
			}

			if utils.HadRuntimeError || utils.HadError {
				t.Fatalf("Unexpected error for %s: %s", tt.name, capturedErr)
			}

			output := strings.TrimSpace(buf.String())
			lines := []string{}
			scannerOut := bufio.NewScanner(strings.NewReader(output))
			for scannerOut.Scan() {
				lines = append(lines, scannerOut.Text())
			}
			if err := scannerOut.Err(); err != nil {
				t.Fatalf("Error reading output: %v", err)
			}

			if len(lines) != len(tt.expected) {
				t.Fatalf("Expected %d lines of output, got %d: %v", len(tt.expected), len(lines), lines)
			}

			for i, exp := range tt.expected {
				if lines[i] != exp {
					t.Errorf("line %d: expected %q, got %q", i+1, exp, lines[i])
				}
			}
		})
	}
}
