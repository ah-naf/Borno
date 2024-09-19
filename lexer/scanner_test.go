package lexer

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

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

func TestScanTokens(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      []token.TokenType
		expectedError string
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: []token.TokenType{token.EOF},
		},
		{
			name:  "Single character tokens",
			input: "(){},.-+;*|&^~",
			expected: []token.TokenType{
				token.LEFT_PAREN, token.RIGHT_PAREN, token.LEFT_BRACE, token.RIGHT_BRACE,
				token.COMMA, token.DOT, token.MINUS, token.PLUS, token.SEMICOLON, token.STAR,
				token.OR, token.AND, token.XOR, token.NOT,
				token.EOF,
			},
		},
		{
			name:  "One and two character tokens",
			input: "! != = == < <= > >= ** << >> || && and or",
			expected: []token.TokenType{
				token.BANG, token.BANG_EQUAL,
				token.EQUAL, token.EQUAL_EQUAL,
				token.LESS, token.LESS_EQUAL,
				token.GREATER, token.GREATER_EQUAL,
				token.POWER,
				token.LEFT_SHIFT, token.RIGHT_SHIFT, token.LOGICAL_OR, token.LOGICAL_AND, token.LOGICAL_AND, token.LOGICAL_OR,
				token.EOF,
			},
		},
		{
			name:     "Comment and whitespace",
			input:    "// This is a comment\n \t\r\n",
			expected: []token.TokenType{token.EOF},
		},
		{
			name:     "String literal",
			input:    `"hello world"`,
			expected: []token.TokenType{token.STRING, token.EOF},
		},
		{
			name:     "Number literal",
			input:    "23 23.32",
			expected: []token.TokenType{token.NUMBER, token.NUMBER, token.EOF},
		},
		{
			name:  "Valid language input",
			input: `var x = 10; print(x + 20);`,
			expected: []token.TokenType{
				token.VAR, token.IDENTIFIER, token.EQUAL, token.NUMBER, token.SEMICOLON,
				token.PRINT, token.LEFT_PAREN, token.IDENTIFIER, token.PLUS, token.NUMBER, token.RIGHT_PAREN, token.SEMICOLON,
				token.EOF,
			},
		},
		{
			name:  "Valid language input with new operators",
			input: `var x = 10; print(x + 20); y = x & 5 | 3 ^ 2; z = x << 2 >> 1;`,
			expected: []token.TokenType{
				token.VAR, token.IDENTIFIER, token.EQUAL, token.NUMBER, token.SEMICOLON,
				token.PRINT, token.LEFT_PAREN, token.IDENTIFIER, token.PLUS, token.NUMBER, token.RIGHT_PAREN, token.SEMICOLON,
				token.IDENTIFIER, token.EQUAL, token.IDENTIFIER, token.AND, token.NUMBER, token.OR, token.NUMBER, token.XOR, token.NUMBER, token.SEMICOLON,
				token.IDENTIFIER, token.EQUAL, token.IDENTIFIER, token.LEFT_SHIFT, token.NUMBER, token.RIGHT_SHIFT, token.NUMBER, token.SEMICOLON,
				token.EOF,
			},
		},
		{
			name:     "Single line comment",
			input:    `// this is a comment`,
			expected: []token.TokenType{token.EOF},
		},
		{
			name:  "Multiline comment in between tokens",
			input: `var x = 10; /* this is a multiline comment */ print(x);`,
			expected: []token.TokenType{
				token.VAR, token.IDENTIFIER, token.EQUAL, token.NUMBER, token.SEMICOLON,
				token.PRINT, token.LEFT_PAREN, token.IDENTIFIER, token.RIGHT_PAREN, token.SEMICOLON,
				token.EOF,
			},
		},
		{
			name:          "Unterminated multiline comment",
			input:         `var x = 10; /* this is an unterminated comment `,
			expected:      []token.TokenType{token.VAR, token.IDENTIFIER, token.EQUAL, token.NUMBER, token.SEMICOLON, token.EOF},
			expectedError: "Unterminated multiline comment",
		},
		{
			name:          "Unterminated string",
			input:         `"This is an unterminated string`,
			expected:      []token.TokenType{token.EOF},
			expectedError: "Unterminated string.",
		},
		{
			name:          "Invalid character",
			input:         "var x = 10; @ print(x);",
			expected:      []token.TokenType{token.VAR, token.IDENTIFIER, token.EQUAL, token.NUMBER, token.SEMICOLON, token.PRINT, token.LEFT_PAREN, token.IDENTIFIER, token.RIGHT_PAREN, token.SEMICOLON, token.EOF},
			expectedError: "Unexpected character.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadError = false

			// Capture stderr output
			capturedErr := CaptureStderr(func() {
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

			splittedErr := strings.Split(capturedErr, "Error: ")
			if len(splittedErr) > 1 {
				capturedErr = strings.TrimSpace(splittedErr[1])
			}

			// Check for expected error messages
			if tt.expectedError != "" {
				if capturedErr == "" {
					t.Errorf("Expected error '%s', but no error was reported", tt.expectedError)
				} else if capturedErr != tt.expectedError {
					t.Errorf("Expected error '%s', but got '%s'", tt.expectedError, capturedErr)
				}
			} else if capturedErr != "" {
				t.Errorf("Did not expect an error, but got: %s", capturedErr)
			}
		})
	}
}
