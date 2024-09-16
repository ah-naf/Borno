package parser_test

import (
	"testing"

	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/lexer"
	"github.com/ah-naf/crafting-interpreter/parser"
)

// Helper function to scan and parse an input expression
func scanAndParse(input string) (ast.Expr, error) {
	// Scan tokens from input using the lexer
	scanner := lexer.NewScanner(input)
	tokens := scanner.ScanTokens()

	// Parse the tokens using the parser
	p := parser.NewParser(tokens)
	expr, err := p.Parse()

	return expr, err
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
			input:     "nil",
			expected:  "nil",
			expectErr: false,
		},
		{
			name:      "Bitwise AND",
			input:     "5 & 3",
			expected:  "(5 & 3)",
			expectErr: false,
		},
		{
			name:      "Bitwise OR",
			input:     "5 | 3",
			expected:  "(5 | 3)",
			expectErr: false,
		},
		{
			name:      "Bitwise XOR",
			input:     "5 ^ 3",
			expected:  "(5 ^ 3)",
			expectErr: false,
		},
		{
			name:      "Bitwise NOT",
			input:     "~5",
			expected:  "(~5)",
			expectErr: false,
		},
		{
			name:      "Left Shift",
			input:     "5 << 2",
			expected:  "(5 << 2)",
			expectErr: false,
		},
		{
			name:      "Right Shift",
			input:     "20 >> 2",
			expected:  "(20 >> 2)",
			expectErr: false,
		},
		{
			name:      "Power",
			input:     "2 ** 3",
			expected:  "(2 ** 3)",
			expectErr: false,
		},
		{
			name:      "Nested grouping",
			input:     "(1 + (2 * 3))",
			expected:  "(group (1 + (group (2 * 3))))",
			expectErr: false,
		},
		{
			name:      "Deeply nested grouping",
			input:     "((1 + 2) * (3 / (4 - 5)))",
			expected:  "(group ((group (1 + 2)) * (group (3 / (group (4 - 5))))))",
			expectErr: false,
		},
		{
			name:      "Chained binary operations",
			input:     "1 + 2 - 3 * 4 / 5",
			expected:  "((1 + 2) - ((3 * 4) / 5))",
			expectErr: false,
		},
		{
			name:      "Complex with comparison",
			input:     "(1 + 2) > (3 * 4) == true",
			expected:  "(((group (1 + 2)) > (group (3 * 4))) == true)",
			expectErr: false,
		},
		{
			name:      "Unary and binary mixed",
			input:     "-(1 + 2) * !(3 > 4)",
			expected:  "((-(group (1 + 2))) * (!(group (3 > 4))))",
			expectErr: false,
		},
		{
			name:      "Complex bitwise operations",
			input:     "(5 & 3) | (4 ^ 2)",
			expected:  "((group (5 & 3)) | (group (4 ^ 2)))",
			expectErr: false,
		},
		{
			name:      "Shift and power",
			input:     "(10 >> 1) ** 2",
			expected:  "((group (10 >> 1)) ** 2)",
			expectErr: false,
		},
		{
			name:      "Precedence test",
			input:     "1 + 2 * 3 ** 2 & 4 | 5 ^ 6",
			expected:  "(((1 + (2 * (3 ** 2))) & 4) | (5 ^ 6))",
			expectErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := scanAndParse(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected an error but got none for input: %s", tt.input)
				}
				return
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Did not expect an error but got one for input: %s, error: %v", tt.input, err)
				return
			}

			if expr.String() != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, expr.String())
			}
		})
	}
}
