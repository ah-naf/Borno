package parser_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/lexer"
	"github.com/ah-naf/crafting-interpreter/parser"
)

// Helper function to scan and parse an input expression
func scanAndParse(input string) ([]ast.Stmt, error) {
	// Scan tokens from input using the lexer
	scanner := lexer.NewScanner(input)
	tokens := scanner.ScanTokens()

	// Parse the tokens using the parser
	p := parser.NewParser(tokens)
	expr, err := p.Parse()

	return expr, err
}

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

// Test cases according to the grammar rules
func TestParseGrammar(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:      "Null",
			input:     "nil;",
			expected:  "nil",
			expectErr: false,
		},
		{
			name:      "Bitwise AND",
			input:     "5 & 3;",
			expected:  "(5 & 3)",
			expectErr: false,
		},
		{
			name:      "Bitwise OR",
			input:     "5 | 3;",
			expected:  "(5 | 3)",
			expectErr: false,
		},
		{
			name:      "Bitwise XOR",
			input:     "5 ^ 3;",
			expected:  "(5 ^ 3)",
			expectErr: false,
		},
		{
			name:      "Bitwise NOT",
			input:     "~5;",
			expected:  "(~5)",
			expectErr: false,
		},
		{
			name:      "Left Shift",
			input:     "5 << 2;",
			expected:  "(5 << 2)",
			expectErr: false,
		},
		{
			name:      "Right Shift",
			input:     "20 >> 2;",
			expected:  "(20 >> 2)",
			expectErr: false,
		},
		{
			name:      "Modulo",
			input:     "20 % 2;",
			expected:  "(20 % 2)",
			expectErr: false,
		},
		{
			name:      "Power",
			input:     "2 ** 3;",
			expected:  "(2 ** 3)",
			expectErr: false,
		},
		{
			name:      "Nested grouping",
			input:     "(1 + (2 * 3));",
			expected:  "(group (1 + (group (2 * 3))))",
			expectErr: false,
		},
		{
			name:      "Deeply nested grouping",
			input:     "((1 + 2) * (3 / (4 - 5)));",
			expected:  "(group ((group (1 + 2)) * (group (3 / (group (4 - 5))))))",
			expectErr: false,
		},
		{
			name:      "Chained binary operations",
			input:     "1 + 2 - 3 * 4 / 5;",
			expected:  "((1 + 2) - ((3 * 4) / 5))",
			expectErr: false,
		},
		{
			name:      "Complex with comparison",
			input:     "(1 + 2) > (3 * 4) == true;",
			expected:  "(((group (1 + 2)) > (group (3 * 4))) == true)",
			expectErr: false,
		},
		{
			name:      "Unary and binary mixed",
			input:     "-(1 + 2) * !(3 > 4);",
			expected:  "((-(group (1 + 2))) * (!(group (3 > 4))))",
			expectErr: false,
		},
		{
			name:      "Complex bitwise operations",
			input:     "(5 & 3) | (4 ^ 2);",
			expected:  "((group (5 & 3)) | (group (4 ^ 2)))",
			expectErr: false,
		},
		{
			name:      "Shift and power",
			input:     "(10 >> 1) ** 2;",
			expected:  "((group (10 >> 1)) ** 2)",
			expectErr: false,
		},
		{
			name:      "Precedence test",
			input:     "1 + 2 * 3 ** 2 & 4 | 5 ^ 6;",
			expected:  "(((1 + (2 * (3 ** 2))) & 4) | (5 ^ 6))",
			expectErr: false,
		},
		{
			name:      "Variable Declaration",
			input:     "var a = 10;",
			expected:  "var a = 10",
			expectErr: false,
		},
		{
			name:      "Multiple Variable Declaration",
			input:     "var a = 10, b = 20;",
			expected:  "var a = 10\nvar b = 20\n",
			expectErr: false,
		},
		{
			name:      "Variable Assignment",
			input:     "a = 10;",
			expected:  "(a = 10)",
			expectErr: false,
		},
		{
			name:      "Invalid Variable Declaration",
			input:     "var = 10;",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Assignment Chaining",
			input:     "var a = b = c = 2;",
			expected:  "var a = (b = (c = 2))",
			expectErr: false,
		},
		{
			name:      "Simple If Statement",
			input:     "if (true) { print 1; }",
			expected:  "if (true){\n(print 1)\n}",
			expectErr: false,
		},
		{
			name:      "If-Else Statement",
			input:     "if (false) { print 1; } else { print 2; }",
			expected:  "if (false){\n(print 1)\n}else {\n(print 2)\n}",
			expectErr: false,
		},
		{
			name:  "Nested If-Else",
			input: "if (a > b) { if (a > c) { print a; } else { print c; } }",
			expected: `if ((a > b)){
if ((a > c)){
(print a)
}else {
(print c)
}
}`,
			expectErr: false,
		},
		{
			name:      "Invalid If Statement",
			input:     "if true { print 1; }",
			expected:  "",
			expectErr: true,
		},
		{
			name:  "valid if statement with && condition",
			input: "if(a > b && a >c ) {print c;} else {print a;}",
			expected: `if (((a > b) && (a > c))){
(print c)
}else {
(print a)
}`,
			expectErr: false,
		},
		{
			name:      "Logical AND condition",
			input:     "if (a > b && b > c) { print a; }",
			expected:  "if (((a > b) && (b > c))){\n(print a)\n}",
			expectErr: false,
		},
		{
			name:      "Logical OR condition",
			input:     "if (a > b || b > c) { print b; }",
			expected:  "if (((a > b) || (b > c))){\n(print b)\n}",
			expectErr: false,
		},
		{
			name:      "Logical AND with Logical OR",
			input:     "if (a > b && b > c || c > d) { print b; }",
			expected:  "if ((((a > b) && (b > c)) || (c > d))){\n(print b)\n}",
			expectErr: false,
		},
		{
			name:      "Nested Logical AND/OR",
			input:     "if ((a > b && b > c) || (c > d && d > e)) { print a; }",
			expected:  "if (((group ((a > b) && (b > c))) || (group ((c > d) && (d > e))))){\n(print a)\n}",
			expectErr: false,
		},
		{
			name:      "Logical OR and comparison",
			input:     "if (a < b || b == c) { print b; }",
			expected:  "if (((a < b) || (b == c))){\n(print b)\n}",
			expectErr: false,
		},
		{
			name:      "Logical AND with comparison and arithmetic",
			input:     "if (a + b > c && b - c < d) { print a; }",
			expected:  "if ((((a + b) > c) && ((b - c) < d))){\n(print a)\n}",
			expectErr: false,
		},
		{
			name:      "Logical AND with nested parentheses",
			input:     "if ((a && b) && (c || d)) { print true; }",
			expected:  "if (((group (a && b)) && (group (c || d)))){\n(print true)\n}",
			expectErr: false,
		},
		{
			name:      "Invalid Logical AND without parentheses",
			input:     "if a && b { print true; }",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Invalid Logical OR without parentheses",
			input:     "if a || b { print false; }",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Simple While Statement",
			input:     "while (true) { print 1; }",
			expected:  "while (true){\n(print 1)\n}",
			expectErr: false,
		},
		{
			name:      "While Statement with Condition",
			input:     "while (x < 5) { x = x + 1; }",
			expected:  "while ((x < 5)){\n(x = (x + 1))\n}",
			expectErr: false,
		},
		{
			name:  "While Statement with Complex Body",
			input: "while (x > 0) { if (x == 1) { print x; } else { print -x; } }",
			expected: `while ((x > 0)){
if ((x == 1)){
(print x)
}else {
(print (-x))
}
}`,
			expectErr: false,
		},
		{
			name: "For loop",
			input: `for(var a = 0; a < 5; a = a + 1) {
				if(a == 2) {
					continue;
				}
				print a;
			}`,
			expected: `for (var a = 0; (a < 5); (a = (a + 1))) {
if ((a == 2)){
continue
}
(print a)
}`,
			expectErr: false,
		},
		{
			name: "Simple For Loop",
			input: `for (var i = 0; i < 10; i = i + 1) { print i; }`,
			expected: `for (var i = 0; (i < 10); (i = (i + 1))) {
(print i)
}`,
			expectErr: false,
		},
		{
			name: "For Loop Without Initializer",
			input: `for (; i < 10; i = i + 1) { print i; }`,
			expected: `for (; (i < 10); (i = (i + 1))) {
(print i)
}`,
			expectErr: false,
		},
		{
			name: "For Loop Without Condition",
			input: `for (var i = 0;; i = i + 1) { print i; }`,
			expected: `for (var i = 0; true; (i = (i + 1))) {
(print i)
}`,
			expectErr: false,
		},
		{
			name: "For Loop Without Increment",
			input: `for (var i = 0; i < 10;) { print i; i = i + 1; }`,
			expected: `for (var i = 0; (i < 10); ) {
(print i)
(i = (i + 1))
}`,
			expectErr: false,
		},
		{
			name: "For Loop Without All Clauses",
			input: `for (;;) { print "infinite"; }`,
			expected: `for (; true; ) {
(print infinite)
}`,
			expectErr: false,
		},
		{
			name: "Invalid For Loop",
			input: `for var i = 0; i < 10; i = i + 1 { print i; }`,
			expected: "",
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := ""
			captured := CaptureStderr(func() {
				expr, err := scanAndParse(tt.input)

				if err != nil {
					os.Stderr.Write([]byte(err.Error() + "\n"))
				}

				if len(expr) == 0 {
					return
				}
				output = expr[0].String()
			})

			if tt.expectErr && captured == "" {
				t.Fatalf("Expected an error but got none for input: %s", tt.input)
			}

			if !tt.expectErr && captured != "" {
				t.Fatalf("Did not expect an error but got one for input: %s, error: %v", tt.input, captured)
			}

			if output != tt.expected {
				t.Fatalf("Expected %s, but got %s", tt.expected, output)
			}

		})
	}
}
