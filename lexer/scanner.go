package lexer

import (
	"strconv"

	"github.com/ah-naf/crafting-interpreter/token"
	"github.com/ah-naf/crafting-interpreter/utils"
)

var keywords = map[string]token.TokenType{
	"and":    token.AND,
	"class":  token.CLASS,
	"else":   token.ELSE,
	"false":  token.FALSE,
	"for":    token.FOR,
	"fun":    token.FUN,
	"if":     token.IF,
	"nil":    token.NIL,
	"or":     token.OR,
	"print":  token.PRINT,
	"return": token.RETURN,
	"super":  token.SUPER,
	"this":   token.THIS,
	"true":   token.TRUE,
	"var":    token.VAR,
	"while":  token.WHILE,
}

type Scanner struct {
	source  string
	tokens  []token.Token
	start   int
	current int
	line    int
}

// NewScanner creates a new Scanner instance
func NewScanner(source string) *Scanner {
	return &Scanner{
		source:  source,
		tokens:  make([]token.Token, 0),
		start:   0,
		current: 0,
		line:    1,
	}
}

// ScanTokens scans the source and returns the list of tokens
func (s *Scanner) ScanTokens() []token.Token {
	for !s.isAtEnd() {
		// We are at the beginning of the next lexeme.
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, *token.NewToken(token.EOF, "", nil, s.line))
	return s.tokens
}

// scanToken scans a single token
func (s *Scanner) scanToken() {
	c := s.advance()

	switch c {
	case '(':
		s.addToken(token.LEFT_PAREN)
	case ')':
		s.addToken(token.RIGHT_PAREN)
	case '{':
		s.addToken(token.LEFT_BRACE)
	case '}':
		s.addToken(token.RIGHT_BRACE)
	case ',':
		s.addToken(token.COMMA)
	case '.':
		s.addToken(token.DOT)
	case '-':
		s.addToken(token.MINUS)
	case '+':
		s.addToken(token.PLUS)
	case ';':
		s.addToken(token.SEMICOLON)
	case '|':
		s.addToken(token.OR)
	case '&':
		s.addToken(token.AND)
	case '^':
		s.addToken(token.XOR)
	case '~':
		s.addToken(token.NEGATION)
	case '*':
		if s.match('*') {
			s.addToken(token.POWER)
		} else {
			s.addToken(token.STAR)
		}
	case '!':
		if s.match('=') {
			s.addToken(token.BANG_EQUAL)
		} else {
			s.addToken(token.BANG)
		}
	case '=':
		if s.match('=') {
			s.addToken(token.EQUAL_EQUAL)
		} else {
			s.addToken(token.EQUAL)
		}
	case '<':
		if s.match('=') {
			s.addToken(token.LESS_EQUAL)
		} else if s.match('<') {
			s.addToken(token.LEFT_SHIFT)
		} else {
			s.addToken(token.LESS)
		}
	case '>':
		if s.match('=') {
			s.addToken(token.GREATER_EQUAL)
		} else if s.match('>') {
			s.addToken(token.RIGHT_SHIFT)
		} else {
			s.addToken(token.GREATER)
		}
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else if s.match('*') {
			s.multilineComment()
		} else {
			s.addToken(token.SLASH)
		}
	case ' ', '\r', '\t':
	case '\n':
		s.line++
	case '"':
		s.stringLiteral()
	default:
		if isDigit(c) {
			s.number()
		} else if isAlpha(c) {
			s.identifier()
		} else {
			utils.GlobalError(s.line, "Unexpected character.")
		}
	}
}

func (s *Scanner) identifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := s.source[s.start:s.current]
	if keyword, ok := keywords[text]; ok {
		s.addToken(keyword)
	} else {
		s.addToken(token.IDENTIFIER)
	}
}

func (s *Scanner) number() {
	for isDigit(s.peek()) {
		s.advance()
	}

	// Look for a fractional part.
	if s.peek() == '.' && isDigit(s.peekNext()) {
		// Consume the "."
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	value, err := strconv.ParseFloat(s.source[s.start:s.current], 64)
	if err != nil {
		utils.GlobalError(s.line, "Invalid number format")
		return
	}

	s.AddToken(token.NUMBER, value)
}

func (s *Scanner) stringLiteral() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		utils.GlobalError(s.line, "Unterminated string.")
		return
	}

	s.advance()

	value := s.source[s.start+1 : s.current-1]
	s.AddToken(token.STRING, value)
}

func (s *Scanner) multilineComment() {
	for !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		} else if s.peek() == '*' && s.peekNext() == '/' {
			// Close the comment
			s.advance() // consume *
			s.advance() // consum /
			return
		}
		s.advance()
	}
	utils.GlobalError(s.line, "Unterminated multiline comment")
}

func (s *Scanner) match(expected byte) bool {
	if s.isAtEnd() {
		return false
	}
	if s.source[s.current] != expected {
		return false
	}
	s.current++
	return true
}

func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return 0
	}
	return s.source[s.current]
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return 0
	}
	return s.source[s.current+1]
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

// isAtEnd checks if we've reached the end of the source
func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) advance() byte {
	b := s.source[s.current]
	s.current++
	return b
}

// addToken adds a new token to the list
func (s *Scanner) addToken(tokenType token.TokenType) {
	s.AddToken(tokenType, nil)
}

func (s *Scanner) AddToken(tokenType token.TokenType, literal interface{}) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, *token.NewToken(tokenType, text, literal, s.line))
}
