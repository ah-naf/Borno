package token

import "fmt"

// TokenType represents the type of a token
type TokenType int

// Token types
const (
	// Single-character tokens
	LEFT_PAREN TokenType = iota
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	LEFT_BRACKET
	RIGHT_BRACKET
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR
	AND
	OR
	XOR
	POWER
	NOT
	MODULO

	// One or two character tokens
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LEFT_SHIFT
	LESS
	LESS_EQUAL
	RIGHT_SHIFT

	// Literals
	IDENTIFIER
	STRING
	NUMBER

	// Keywords
	BREAK
	CONTINUE
	LOGICAL_AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	LOGICAL_OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE

	EOF
)

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal interface{}
	Line    int
}

// NewToken creates a new Token instance
func NewToken(tokenType TokenType, lexeme string, literal interface{}, line int) *Token {
	return &Token{
		Type:    tokenType,
		Lexeme:  lexeme,
		Literal: literal,
		Line:    line,
	}
}

// String returns a string representation of the Token
func (t *Token) String() string {
	return fmt.Sprintf("%v %s %v", t.Type, t.Lexeme, t.Literal)
}
