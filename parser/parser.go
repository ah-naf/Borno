package parser

import (
	"fmt"

	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/token"
	"github.com/ah-naf/crafting-interpreter/utils"
)

type ParseError struct {
	message string
}

func (e ParseError) Error() string {
	return e.message
}

type Parser struct {
	tokens  []token.Token
	current int
}

func NewParser(tokens []token.Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() (ast.Expr, error) {
	return p.expression()
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.equality()
}

func (p *Parser) equality() (ast.Expr, error) {
	expr, err := p.comparison()

	if err != nil {
		return nil, err
	}

	for p.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.previous()
		right, err := p.comparison()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) comparison() (ast.Expr, error) {
	expr, err := p.term()

	if err != nil {
		return nil, err
	}

	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right, err := p.term()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) term() (ast.Expr, error) {
	expr, err := p.factor()

	if err != nil {
		return nil, err
	}

	for p.match(token.MINUS, token.PLUS) {
		operator := p.previous()
		right, err := p.factor()

		

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) factor() (ast.Expr, error) {
	expr, err := p.unary()

	if err != nil {
		return nil, err
	}

	for p.match(token.SLASH, token.STAR) {
		operator := p.previous()
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) unary() (ast.Expr, error) {
	if p.match(token.BANG, token.MINUS) {
		operator := p.previous()
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		return &ast.Unary{Operator: operator, Right: right}, nil
	}

	return p.primary()
}

func (p *Parser) primary() (ast.Expr, error) {
	if p.match(token.FALSE) {
		return &ast.Literal{Value: false}, nil
	}
	if p.match(token.TRUE) {
		return &ast.Literal{Value: true}, nil
	}
	if p.match(token.NIL) {
		return &ast.Literal{Value: nil}, nil
	}

	if p.match(token.NUMBER, token.STRING) {
		return &ast.Literal{Value: p.previous().Literal}, nil
	}

	if p.match(token.LEFT_PAREN) {
		expr, err := p.expression()
		
		if err != nil {
			return nil, err
		}

		_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after expression.")

		if err != nil {
			return nil, err
		}

		return &ast.Grouping{Expression: expr}, nil
	}

	return nil, p.error(p.peek(), "Unexpected token. Expect expression.")
}

func (p *Parser) match(types ...token.TokenType) bool {
	for _, tt := range types {
		if p.check(tt) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) consume(tokenType token.TokenType, message string) (token.Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}

	return token.Token{}, p.error(p.peek(), message)
}

func (p *Parser) error(t token.Token, message string) error {
	utils.GlobalErrorToken(t, message)
	return fmt.Errorf(message)
}

func (p *Parser) check(tokenType token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tokenType
}

func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}

func (p *Parser) peek() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.current-1]
}
