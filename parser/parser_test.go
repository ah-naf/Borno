package parser_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ah-naf/borno/ast"
	"github.com/ah-naf/borno/lexer"
	"github.com/ah-naf/borno/parser"
)

// Helper function to scan and parse an input expression
func scanAndParse(input string) ([]ast.Stmt, error) {
	// Scan tokens from input using the lexer
	inputRune := []rune(input)
	scanner := lexer.NewScanner(inputRune)
	tokens := scanner.ScanTokens()

	// Parse the tokens using the parser
	p := parser.NewParser(tokens)
	expr, err := p.Parse()

	return expr, err
}

func CaptureStderr(f func()) string {
	// Create a pipe to capture os.Stderr
	r, w, _ := os.Pipe()
	oldStderr := os.Stderr
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

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
			input:     "(1 + 2) > (3 * 4) == সত্য;",
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
			input:     "ধরি a = 10;",
			expected:  "var a = 10",
			expectErr: false,
		},
		{
			name:      "Multiple Variable Declaration",
			input:     "ধরি a = 10, b = 20;",
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
			input:     "ধরি = 10;",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Assignment Chaining",
			input:     "ধরি a = b = c = 2;",
			expected:  "var a = (b = (c = 2))",
			expectErr: false,
		},
		{
			name:      "Simple If Statement",
			input:     "যদি (সত্য) { দেখাও 1; }",
			expected:  "if (true){\n(print 1)\n}",
			expectErr: false,
		},
		{
			name:      "If-Else Statement",
			input:     "যদি (মিথ্যা) { দেখাও 1; } নাহয় { দেখাও 2; }",
			expected:  "if (false){\n(print 1)\n}else {\n(print 2)\n}",
			expectErr: false,
		},
		{
			name:  "Nested If-Else",
			input: "যদি (a > b) { যদি (a > c) { দেখাও a; } নাহয় { দেখাও c; } }",
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
			input:     "যদি সত্য { দেখাও 1; }",
			expected:  "",
			expectErr: true,
		},
		{
			name:  "valid if statement with && condition",
			input: "যদি(a > b && a > c ) {দেখাও c;} নাহয় {দেখাও a;}",
			expected: `if (((a > b) && (a > c))){
(print c)
}else {
(print a)
}`,
			expectErr: false,
		},
		{
			name:      "Logical AND condition",
			input:     "যদি (a > b && b > c) { দেখাও a; }",
			expected:  "if (((a > b) && (b > c))){\n(print a)\n}",
			expectErr: false,
		},
		{
			name:      "Logical OR condition",
			input:     "যদি (a > b || b > c) { দেখাও b; }",
			expected:  "if (((a > b) || (b > c))){\n(print b)\n}",
			expectErr: false,
		},
		{
			name:      "Logical AND with Logical OR",
			input:     "যদি (a > b && b > c || c > d) { দেখাও b; }",
			expected:  "if ((((a > b) && (b > c)) || (c > d))){\n(print b)\n}",
			expectErr: false,
		},
		{
			name:      "Nested Logical AND/OR",
			input:     "যদি ((a > b && b > c) || (c > d && d > e)) { দেখাও a; }",
			expected:  "if (((group ((a > b) && (b > c))) || (group ((c > d) && (d > e))))){\n(print a)\n}",
			expectErr: false,
		},
		{
			name:      "Logical OR and comparison",
			input:     "যদি (a < b || b == c) { দেখাও b; }",
			expected:  "if (((a < b) || (b == c))){\n(print b)\n}",
			expectErr: false,
		},
		{
			name:      "Logical AND with comparison and arithmetic",
			input:     "যদি (a + b > c && b - c < d) { দেখাও a; }",
			expected:  "if ((((a + b) > c) && ((b - c) < d))){\n(print a)\n}",
			expectErr: false,
		},
		{
			name:      "Logical AND with nested parentheses",
			input:     "যদি ((a && b) && (c || d)) { দেখাও সত্য; }",
			expected:  "if (((group (a && b)) && (group (c || d)))){\n(print true)\n}",
			expectErr: false,
		},
		{
			name:      "Invalid Logical AND without parentheses",
			input:     "যদি a && b { দেখাও সত্য; }",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Invalid Logical OR without parentheses",
			input:     "যদি a || b { দেখাও মিথ্যা; }",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Simple While Statement",
			input:     "যতক্ষণ (সত্য) { দেখাও 1; }",
			expected:  "while (true){\n(print 1)\n}",
			expectErr: false,
		},
		{
			name:      "While Statement with Condition",
			input:     "যতক্ষণ (x < 5) { x = x + 1; }",
			expected:  "while ((x < 5)){\n(x = (x + 1))\n}",
			expectErr: false,
		},
		{
			name:  "While Statement with Complex Body",
			input: "যতক্ষণ (x > 0) { যদি (x == 1) { দেখাও x; } নাহয় { দেখাও -x; } }",
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
			input: `ফর(ধরি a = 0; a < 5; a = a + 1) {
						যদি(a == 2) {
							চালিয়ে_যাও;
						}
						দেখাও a;
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
			name:  "Simple For Loop",
			input: `ফর (ধরি i = 0; i < 10; i = i + 1) { দেখাও i; }`,
			expected: `for (var i = 0; (i < 10); (i = (i + 1))) {
(print i)
}`,
			expectErr: false,
		},
		{
			name:  "For Loop Without Initializer",
			input: `ফর (; i < 10; i = i + 1) { দেখাও i; }`,
			expected: `for (; (i < 10); (i = (i + 1))) {
(print i)
}`,
			expectErr: false,
		},
		{
			name:  "For Loop Without Condition",
			input: `ফর (ধরি i = 0;; i = i + 1) { দেখাও i; }`,
			expected: `for (var i = 0; true; (i = (i + 1))) {
(print i)
}`,
			expectErr: false,
		},
		{
			name:  "For Loop Without Increment",
			input: `ফর (ধরি i = 0; i < 10;) { দেখাও i; i = i + 1; }`,
			expected: `for (var i = 0; (i < 10); ) {
(print i)
(i = (i + 1))
}`,
			expectErr: false,
		},
		{
			name:  "For Loop Without All Clauses",
			input: `ফর (;;) { দেখাও "infinite"; }`,
			expected: `for (; true; ) {
(print infinite)
}`,
			expectErr: false,
		},
		{
			name:      "Invalid For Loop",
			input:     `ফর ধরি i = 0; i < 10; i = i + 1 { দেখাও i; }`,
			expected:  "",
			expectErr: true,
		},
		{
			name: "Basic Function",
			input: `ফাংশন add(a, b) {
দেখাও a + b;
}`,
			expected: `fun add(a, b) {
(print (a + b))
}`,
			expectErr: false,
		},
		{
			name:      "Basic Function Call",
			input:     `add(a, b);`,
			expected:  `add(a, b)`,
			expectErr: false,
		},
		{
			name:      "Function Call Error",
			input:     `add(a,);`,
			expected:  ``,
			expectErr: true,
		},
		{
			name: "Simple Function Declaration",
			input: `ফাংশন sayHello() {
দেখাও "Hello!";
}`,
			expected: `fun sayHello() {
(print Hello!)
}`,
			expectErr: false,
		},
		{
			name:      "Function Call Without Arguments",
			input:     `sayHello();`,
			expected:  `sayHello()`,
			expectErr: false,
		},
		{
			name:      "Function Call With Arguments",
			input:     `add(5, 10);`,
			expected:  `add(5, 10)`,
			expectErr: false,
		},
		{
			name:      "Invalid Function Call with Trailing Comma",
			input:     `add(5,);`,
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Nested Function Call",
			input:     `add(multiply(2, 3), 5);`,
			expected:  `add(multiply(2, 3), 5)`,
			expectErr: false,
		},
		{
			name:      "Return statemetn",
			input:     "ফেরত a;",
			expected:  "return a",
			expectErr: false,
		},
		{
			name:      "Array Literal",
			input:     "ধরি arr = [1, 2, 3];",
			expected:  "var arr = [1, 2, 3]",
			expectErr: false,
		},
		{
			name:      "Empty Array Literal",
			input:     "ধরি arr = [];",
			expected:  "var arr = []",
			expectErr: false,
		},
		{
			name:      "Array Access",
			input:     "arr[0];",
			expected:  "arr[0]",
			expectErr: false,
		},
		{
			name:      "Nested Array Access",
			input:     "arr[0][1];",
			expected:  "arr[0][1]",
			expectErr: false,
		},
		{
			name:      "Array in Expression",
			input:     "দেখাও arr[0] + 5;",
			expected:  "(print (arr[0] + 5))",
			expectErr: false,
		},
		{
			name:      "Invalid Array Access",
			input:     "arr[;",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Invalid Array Literal",
			input:     "ধরি arr = [1, 2, ;",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Array Assignment",
			input:     "arr[0] = 10;",
			expected:  "(arr[0] = 10)",
			expectErr: false,
		},
		{
			name:      "Array Funtion call",
			input:     "arr[0]();",
			expected:  "arr[0]()",
			expectErr: false,
		},
		{
			name:      "Object Literal",
			input:     `ধরি obj = {name: "Alice", age: 30, height: 5.9};`,
			expected:  `var obj = {name: Alice, age: 30, height: 5.9}`,
			expectErr: false,
		},
		{
			name:      "Object with Numeric Keys",
			input:     `ধরি obj = {1: "one", 2: "two"};`,
			expected:  ``,
			expectErr: true,
		},
		{
			name:      "Empty Object Literal",
			input:     `ধরি obj = {};`,
			expected:  `var obj = {}`,
			expectErr: false,
		},
		{
			name:      "Object Property Access",
			input:     `obj.name;`,
			expected:  `obj.name`,
			expectErr: false,
		},
		{
			name:      "Object Property Assignment",
			input:     `obj.name = "Bob";`,
			expected:  `obj.name = Bob`,
			expectErr: false,
		},
		{
			name:      "Nested Object Access",
			input:     `person.address.street;`,
			expected:  `person.address.street`,
			expectErr: false,
		},
		{
			name:      "Nested Object Assignment",
			input:     `person.address.street = "Main St";`,
			expected:  `person.address.street = Main St`,
			expectErr: false,
		},
		{
			name:      "Invalid Property Access",
			input:     `obj.;`,
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Object Property Access with Array",
			input:     `person.children[0].name;`,
			expected:  `person.children[0].name`,
			expectErr: false,
		},
		{
			name:      "Object Property Assignment with Array",
			input:     `person.children[0].name = "Charlie";`,
			expected:  `person.children[0].name = Charlie`,
			expectErr: false,
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
