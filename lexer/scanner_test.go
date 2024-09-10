package lexer

import (
	"testing"

	token "github.com/ah-naf/crafting-interpreter/token"
)

func TestScanTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []token.TokenType
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: []token.TokenType{token.EOF},
		},
		{
			name:  "Single character tokens",
			input: "(){},.-+;*",
			expected: []token.TokenType{
				token.LEFT_PAREN, token.RIGHT_PAREN, token.LEFT_BRACE, token.RIGHT_BRACE,
				token.COMMA, token.DOT, token.MINUS, token.PLUS, token.SEMICOLON, token.STAR,
				token.EOF,
			},
		},
		{
			name:  "One and two character tokens",
			input: "! != = == < <= > >=",
			expected: []token.TokenType{
				token.BANG, token.BANG_EQUAL,
				token.EQUAL, token.EQUAL_EQUAL,
				token.LESS, token.LESS_EQUAL,
				token.GREATER, token.GREATER_EQUAL,
				token.EOF,
			},
		},
		{
			name:     "Comment and whitespace",
			input:    "// This is a comment\n  \t\r\n",
			expected: []token.TokenType{token.EOF},
		},
		{
			name:  "String literal",
			input: `"hello world"`,
			expected: []token.TokenType{
				token.STRING, token.EOF,
			},
		},
		{
			name:  "Number literal",
			input: "23 23.32",
			expected: []token.TokenType{
				token.NUMBER, token.NUMBER, token.EOF,
			},
		},
		{
			name:  "Valid language input",
			input: `var x = 10; print(x + 20);`,
			expected: []token.TokenType{
				token.VAR, token.IDENTIFIER, token.EQUAL, token.NUMBER, token.SEMICOLON, // var x = 10;
				token.PRINT, token.LEFT_PAREN, token.IDENTIFIER, token.PLUS, token.NUMBER, token.RIGHT_PAREN, token.SEMICOLON, // print(x + 20);
				token.EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner(tt.input)
			tokens := scanner.ScanTokens()
			if len(tokens) != len(tt.expected) {
				t.Errorf("Expected %d tokens, but got %d", len(tt.expected), len(tokens))
				return
			}

			for i, expectedType := range tt.expected {
				if tokens[i].Type != expectedType {
					t.Errorf("Token %d: expected %v, but got %v", i, expectedType, tokens[i].Type)
				}
			}
		})
	}
}
