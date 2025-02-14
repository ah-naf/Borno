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

func TestScanTokensBanglaKeywords(t *testing.T) {
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
			name: "Single character tokens",
			// Punctuation & operators that do not rely on keywords
			input: "(){}[],.-+;*|&^~%:",
			expected: []token.TokenType{
				token.LEFT_PAREN, token.RIGHT_PAREN,
				token.LEFT_BRACE, token.RIGHT_BRACE,
				token.LEFT_BRACKET, token.RIGHT_BRACKET,
				token.COMMA, token.DOT, token.MINUS, token.PLUS,
				token.SEMICOLON, token.STAR,
				token.OR, token.AND, token.XOR, token.NOT,
				token.MODULO, token.COLON,
				token.EOF,
			},
		},
		{
			name:     "Number literal",
			input:    "২৩ ২৩.৩২",
			expected: []token.TokenType{token.NUMBER, token.NUMBER, token.EOF},
		},
		{
			name:  "Variable declaration and print",
			input: `ধরি x = 10; দেখাও(x + 20);`,
			expected: []token.TokenType{
				token.VAR,         // "ধরি"
				token.IDENTIFIER,  // "x"
				token.EQUAL,       // '='
				token.NUMBER,      // "10"
				token.SEMICOLON,   // ';'
				token.PRINT,       // "print"
				token.LEFT_PAREN,  // '('
				token.IDENTIFIER,  // "x"
				token.PLUS,        // '+'
				token.NUMBER,      // "20"
				token.RIGHT_PAREN, // ')'
				token.SEMICOLON,   // ';'
				token.EOF,
			},
		},
		{
			name: "If-else statement in Bangla",
			// যদি (x > ১০) { print("বড়"); } নাহয় { print("ছোট"); }
			input: `যদি (x > 10) { দেখাও("বড়"); } নাহয় { দেখাও("ছোট"); }`,
			expected: []token.TokenType{
				token.IF,          // "যদি"
				token.LEFT_PAREN,  // '('
				token.IDENTIFIER,  // "x"
				token.GREATER,     // '>'
				token.NUMBER,      // "10"
				token.RIGHT_PAREN, // ')'
				token.LEFT_BRACE,  // '{'
				token.PRINT,       // "print"
				token.LEFT_PAREN,  // '('
				token.STRING,      // "বড়"
				token.RIGHT_PAREN, // ')'
				token.SEMICOLON,   // ';'
				token.RIGHT_BRACE, // '}'
				token.ELSE,        // "নাহয়"
				token.LEFT_BRACE,  // '{'
				token.PRINT,       // "print"
				token.LEFT_PAREN,  // '('
				token.STRING,      // "ছোট"
				token.RIGHT_PAREN, // ')'
				token.SEMICOLON,   // ';'
				token.RIGHT_BRACE, // '}'
				token.EOF,
			},
		},
		{
			name: "Function definition in Bangla",
			// ফাংশন greet(নাম) { print("হ্যালো, " + নাম); }
			input: `ফাংশন greet(নাম) { দেখাও("হ্যালো, " + নাম); }`,
			expected: []token.TokenType{
				token.FUN,         // "ফাংশন"
				token.IDENTIFIER,  // "greet"
				token.LEFT_PAREN,  // '('
				token.IDENTIFIER,  // "নাম"
				token.RIGHT_PAREN, // ')'
				token.LEFT_BRACE,  // '{'
				token.PRINT,       // "print"
				token.LEFT_PAREN,  // '('
				token.STRING,      // "হ্যালো, "
				token.PLUS,        // '+'
				token.IDENTIFIER,  // "নাম"
				token.RIGHT_PAREN, // ')'
				token.SEMICOLON,   // ';'
				token.RIGHT_BRACE, // '}'
				token.EOF,
			},
		},
		{
			name: "While statement in Bangla",
			// যতক্ষণ (x < 10) { x = x + 1; }
			input: `যতক্ষণ (x < 10) { x = x + 1; }`,
			expected: []token.TokenType{
				token.WHILE,       // "যতক্ষণ"
				token.LEFT_PAREN,  // '('
				token.IDENTIFIER,  // "x"
				token.LESS,        // '<'
				token.NUMBER,      // "10"
				token.RIGHT_PAREN, // ')'
				token.LEFT_BRACE,  // '{'
				token.IDENTIFIER,  // "x"
				token.EQUAL,       // '='
				token.IDENTIFIER,  // "x"
				token.PLUS,        // '+'
				token.NUMBER,      // "1"
				token.SEMICOLON,   // ';'
				token.RIGHT_BRACE, // '}'
				token.EOF,
			},
		},
		{
			name: "Return / Break / Continue in Bangla",
			// ফেরত ১০; থামো; চালিয়ে_যাও;
			input: `ফেরত 10; থামো; চালিয়ে_যাও;`,
			expected: []token.TokenType{
				token.RETURN,    // "ফেরত"
				token.NUMBER,    // "10"
				token.SEMICOLON, // ';'
				token.BREAK,     // "থামো"
				token.SEMICOLON, // ';'
				token.CONTINUE,  // "চালিয়ে_যাও"
				token.SEMICOLON, // ';'
				token.EOF,
			},
		},
		{
			name: "Logical operators in Bangla",
			// (সত্য এবং মিথ্যা) বা মিথ্যা
			// Depending on your grammar: "সত্য", "মিথ্যা", "এবং", "বা"
			input: `(সত্য এবং মিথ্যা) বা মিথ্যা`,
			expected: []token.TokenType{
				token.LEFT_PAREN,  // '('
				token.TRUE,        // "সত্য"
				token.LOGICAL_AND, // "এবং"
				token.FALSE,       // "মিথ্যা"
				token.RIGHT_PAREN, // ')'
				token.LOGICAL_OR,  // "বা"
				token.FALSE,       // "মিথ্যা"
				token.EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.HadError = false

			// Capture stderr output
			capturedErr := CaptureStderr(func() {
				// Pass the Bangla input as runes to NewScanner
				scanner := NewScanner([]rune(tt.input))
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

			// Grab the actual error message (if any)
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
