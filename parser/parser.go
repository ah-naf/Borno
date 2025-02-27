package parser

import (
	"fmt"

	"github.com/ah-naf/borno/ast"
	"github.com/ah-naf/borno/token"
	"github.com/ah-naf/borno/utils"
)

var reservedIdentifiers = map[string]bool{
	"ক্লক":         true,
	"লেন":          true,
	"এড":           true,
	"রিমুভ":        true,
	"কি_রিমুভ":     true,
	"অব্জেক্ট_কি":  true,
	"অব্জেক্ট_মান": true,
	"পরমমান":       true,
	"বর্গমূল":      true,
	"ঘাত":          true,
	"সাইন":         true,
	"কসাইন":        true,
	"ট্যান":        true,
	"সর্বনিম্ন":    true,
	"সর্বোচ্চ":     true,
	"রাউন্ড":       true,
	"input":        true,
	"ইনপুট":        true,
}

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

func (p *Parser) Parse() ([]ast.Stmt, error) {
	statments := []ast.Stmt{}

	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statments = append(statments, stmt)
	}

	return statments, nil
}

func (p *Parser) declaration() (ast.Stmt, error) {
	if p.match(token.FUN) {
		return p.function("function")
	}
	if p.match(token.VAR) {
		return p.varDeclaration()
	}
	return p.statement()
}

func (p *Parser) varDeclaration() (ast.Stmt, error) {
	var declarations []ast.VarStmt
	initialLine := p.peek().Line // Track the line number at the start of the declaration

	for {
		// Parse the variable name
		name, err := p.consume(token.IDENTIFIER, "Expect variable name.")
		if err != nil {
			return nil, err
		}

		// Check if the name is a reserved identifier
		if _, isReserved := reservedIdentifiers[name.Lexeme]; isReserved {
			return nil, p.error(name, fmt.Sprintf("'%s' is a reserved identifier and cannot be used as a variable name.", name.Lexeme))
		}

		// Optional initializer
		var initializer ast.Expr
		if p.match(token.EQUAL) {
			val, err := p.expression()
			if err != nil {
				return nil, err
			}
			initializer = val
		}

		// Create a VarStmt for each variable
		declaration := &ast.VarStmt{Name: name, Initializer: initializer, Line: name.Line}
		declarations = append(declarations, *declaration)

		// Check for newline and semicolon before proceeding to the next variable,
		// but skip this check if the initializer is an object or array literal.
		switch initializer.(type) {
		case *ast.ObjectLiteral, *ast.ArrayLiteral:
			// Skip newline check for ObjectLiteral and ArrayLiteral
		default:
			if p.peek().Line != initialLine {
				return nil, p.error(p.peek(), "Expect ';' before newline.")
			}
		}

		// If no more commas, break out of the loop
		if !p.match(token.COMMA) {
			break
		}
	}

	// Ensure semicolon at the end of the declaration
	_, err := p.consume(token.SEMICOLON, "Expect ';' after variable declaration.")
	if err != nil {
		return nil, err
	}

	// If there's only one variable, return it directly
	if len(declarations) == 1 {
		return &declarations[0], nil
	}

	// If there are multiple variables, return a VarListStmt
	return &ast.VarListStmt{Declarations: declarations}, nil
}

func (p *Parser) statement() (ast.Stmt, error) {
	if p.match(token.IF) {
		return p.IfStatement()
	}
	if p.match(token.WHILE) {
		return p.while()
	}
	if p.match(token.FOR) {
		return p.forStatement()
	}
	if p.match(token.PRINT) {
		return p.printStatement()
	}
	if p.match(token.RETURN) {
		return p.returnStatement()
	}
	if p.match(token.BREAK) {
		_, err := p.consume(token.SEMICOLON, "Expected ; after break.")
		if err != nil {
			return nil, err
		}
		return &ast.BreakStmt{Line: p.previous().Line}, nil
	}
	if p.match(token.CONTINUE) {
		_, err := p.consume(token.SEMICOLON, "Expected ; after continue.")
		if err != nil {
			return nil, err
		}
		return &ast.ContinueStmt{Line: p.previous().Line}, nil
	}

	if p.match(token.LEFT_BRACE) {
		blocks, err := p.block()
		if err != nil {
			return nil, err
		}
		return &ast.BlockStmt{Block: blocks}, nil
	}

	return p.expressionStatement()
}

func (p *Parser) forStatement() (ast.Stmt, error) {
	_, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'for'.")
	if err != nil {
		return nil, err
	}

	var initializer ast.Stmt
	if p.match(token.SEMICOLON) {
		initializer = nil
	} else if p.match(token.VAR) {
		initializer, err = p.varDeclaration()
		if err != nil {
			return nil, err
		}
	} else {
		initializer, err = p.expressionStatement()
		if err != nil {
			return nil, err
		}
	}
	var condition ast.Expr
	if !p.check(token.SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after loop condition.")
	if err != nil {
		return nil, err
	}

	var increment ast.Expr
	if !p.check(token.RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after for clauses.")
	if err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	if condition == nil {
		condition = &ast.Literal{Value: true}
	}

	return &ast.ForStmt{Initializer: initializer, Condition: condition, Body: body, Increment: increment}, nil
}

func (p *Parser) while() (ast.Stmt, error) {
	_, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'while'.")
	if err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after condition.")
	if err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return &ast.While{Condition: condition, Body: body}, nil
}

func (p *Parser) IfStatement() (ast.Stmt, error) {
	_, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'if'.")
	if err != nil {
		return nil, err
	}
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after if condition.")
	if err != nil {
		return nil, err
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}
	var elseBranch ast.Stmt
	if p.match(token.ELSE) {
		v, err := p.statement()
		if err != nil {
			return nil, err
		}
		elseBranch = v
	}
	return &ast.IfStmt{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}, nil
}

func (p *Parser) printStatement() (ast.Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}
	p.consume(token.SEMICOLON, "Expect ';' after value.")
	return &ast.PrintStatement{Expression: value}, nil
}

func (p *Parser) returnStatement() (ast.Stmt, error) {
	keyword := p.previous()
	var value ast.Expr

	if !p.check(token.SEMICOLON) {
		v, err := p.expression()
		if err != nil {
			return nil, err
		}
		value = v
	}

	_, err := p.consume(token.SEMICOLON, "Expect ';' after return value.")
	if err != nil {
		return nil, err
	}

	return &ast.Return{Keyword: keyword, Value: value}, nil
}

func (p *Parser) expressionStatement() (ast.Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}
	p.consume(token.SEMICOLON, "Expect ';' after value.")
	return &ast.ExpressionStatement{Expression: value}, nil
}

func (p *Parser) function(kind string) (ast.Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect "+kind+" name.")
	if err != nil {
		return nil, err
	}

	if _, isReserved := reservedIdentifiers[name.Lexeme]; isReserved {
		return nil, p.error(name, fmt.Sprintf("'%s' is a reserved identifier and cannot be used as a function name.", name.Lexeme))
	}

	_, err = p.consume(token.LEFT_PAREN, "Expect '(' after "+kind+" name.")
	if err != nil {
		return nil, err
	}

	parameters := []token.Token{}
	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(parameters) >= 255 {
				return nil, p.error(p.peek(), "Can't have more than 255 parameters.")
			}

			pp, err := p.consume(token.IDENTIFIER, "Expect parameter name.")
			if err != nil {
				return nil, err
			}
			parameters = append(parameters, pp)

			if !p.match(token.COMMA) {
				break
			}
		}
	}
	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after parameters.")
	if err != nil {
		return nil, err
	}

	_, err = p.consume(token.LEFT_BRACE, "Expect '{' before "+kind+" body.")
	if err != nil {
		return nil, err
	}

	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return &ast.FunctionStmt{Name: name, Params: parameters, Body: body}, nil
}

func (p *Parser) block() ([]ast.Stmt, error) {
	statments := []ast.Stmt{}

	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		decl, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statments = append(statments, decl)
	}

	p.consume(token.RIGHT_BRACE, "Expect '}' after block.")
	return statments, nil
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (ast.Expr, error) {
	// Parse the expression on the left-hand side of the assignment
	expr, err := p.logicalOR()
	if err != nil {
		return nil, err
	}

	// Check if the current token is an assignment operator
	if p.match(token.EQUAL) {
		equalOperator := p.previous()

		// Parse the expression on the right-hand side of the assignment
		value, err := p.assignment()
		if err != nil {
			return nil, err
		}

		// fmt.Printf("%#v\n", expr)
		// Ensure that the left-hand side is a valid assignment target
		switch target := expr.(type) {
		case *ast.Identifier:
			// If the left-hand side is an identifier, it's a valid assignment target
			return &ast.AssignmentStmt{
				Name:  target.Name,
				Value: value,
				Line:  equalOperator.Line,
			}, nil
		case *ast.ArrayAccess:
			// If the left-hand side is an array access, it's also a valid assignment target
			return &ast.ArrayAssignment{
				Array: target.Array,
				Index: target.Index,
				Value: value,
				Line:  equalOperator.Line,
			}, nil
		case *ast.PropertyAccess:
			// Handle object property access assignment
			return &ast.PropertyAssignment{
				Object:   target.Object,
				Property: target.Property,
				Value:    value,
				Line:     equalOperator.Line,
			}, nil
		default:
			// If the left-hand side is neither, throw an error
			return nil, p.error(equalOperator, "Invalid assignment target.")
		}
	}

	// If no assignment, return the original expression
	return expr, nil
}

func (p *Parser) logicalOR() (ast.Expr, error) {
	expr, err := p.logicalAnd()
	if err != nil {
		return nil, err
	}

	for p.match(token.LOGICAL_OR) {
		operator := p.previous()
		right, err := p.logicalAnd()
		if err != nil {
			return nil, err
		}

		expr = &ast.Logical{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) logicalAnd() (ast.Expr, error) {
	expr, err := p.bitwiseOR()
	if err != nil {
		return nil, err
	}

	for p.match(token.LOGICAL_AND) {
		operator := p.previous()
		right, err := p.bitwiseOR()
		if err != nil {
			return nil, err
		}

		expr = &ast.Logical{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) bitwiseOR() (ast.Expr, error) {
	expr, err := p.bitwiseXOR()

	if err != nil {
		return nil, err
	}

	for p.match(token.OR) {
		operator := p.previous()
		right, err := p.bitwiseXOR()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right, Line: operator.Line}
	}

	return expr, nil
}

func (p *Parser) bitwiseXOR() (ast.Expr, error) {
	expr, err := p.bitwiseAND()

	if err != nil {
		return nil, err
	}

	for p.match(token.XOR) {
		operator := p.previous()
		right, err := p.bitwiseAND()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right, Line: operator.Line}
	}
	return expr, nil
}

func (p *Parser) bitwiseAND() (ast.Expr, error) {
	expr, err := p.equality()

	if err != nil {
		return nil, err
	}

	for p.match(token.AND) {
		operator := p.previous()
		right, err := p.equality()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right, Line: operator.Line}
	}
	return expr, nil
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

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right, Line: operator.Line}
	}

	return expr, nil
}

func (p *Parser) comparison() (ast.Expr, error) {
	expr, err := p.shift()

	if err != nil {
		return nil, err
	}

	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right, err := p.shift()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right, Line: operator.Line}
	}

	return expr, nil
}

func (p *Parser) shift() (ast.Expr, error) {
	expr, err := p.term()

	if err != nil {
		return nil, err
	}

	for p.match(token.LEFT_SHIFT, token.RIGHT_SHIFT) {
		operator := p.previous()
		right, err := p.term()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right, Line: operator.Line}
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

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right, Line: operator.Line}
	}

	return expr, nil
}

func (p *Parser) factor() (ast.Expr, error) {
	expr, err := p.power()

	if err != nil {
		return nil, err
	}

	for p.match(token.SLASH, token.STAR, token.MODULO) {
		operator := p.previous()
		right, err := p.power()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right, Line: operator.Line}
	}

	return expr, nil
}

func (p *Parser) power() (ast.Expr, error) {
	expr, err := p.unary()

	if err != nil {
		return nil, err
	}

	for p.match(token.POWER) {
		operator := p.previous()
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{Left: expr, Operator: operator, Right: right, Line: operator.Line}
	}

	return expr, nil
}

func (p *Parser) unary() (ast.Expr, error) {
	if p.match(token.BANG, token.MINUS, token.NOT) {
		operator := p.previous()
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		return &ast.Unary{Operator: operator, Right: right, Line: operator.Line}, nil
	}

	return p.call()
}

func (p *Parser) call() (ast.Expr, error) {
	// Start by parsing the primary expression (the callee).
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	// Continue to check for function calls (which may be chained).
	for {
		if p.match(token.LEFT_PAREN) {
			// If the next token is '(', finish parsing the call expression.
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(token.LEFT_BRACKET) {
			index, err := p.expression()
			if err != nil {
				return nil, err
			}

			_, err = p.consume(token.RIGHT_BRACKET, "Expect ']' after array index.")
			if err != nil {
				return nil, err
			}
			expr = &ast.ArrayAccess{Array: expr, Index: index, Line: p.previous().Line}
			// fmt.Printf("%#v\n", expr)
		} else if p.match(token.DOT) {
			// Handle property access
			propName, err := p.consume(token.IDENTIFIER, "Expect property name after '.'.")
			if err != nil {
				return nil, err
			}
			expr = &ast.PropertyAccess{Object: expr, Property: propName, Line: p.previous().Line}
		} else {
			break // No more call expressions to parse.
		}
	}
	return expr, nil
}

func (p *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	// Parse the arguments inside the parentheses.
	arguments := []ast.Expr{}

	if !p.check(token.RIGHT_PAREN) { // If there are arguments to parse.
		for {
			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, arg)

			// Continue parsing arguments separated by commas.
			if !p.match(token.COMMA) {
				break
			}
		}
	}

	// Ensure the call expression ends with a closing parenthesis.
	paren, err := p.consume(token.RIGHT_PAREN, "Expect ')' after arguments.")
	if err != nil {
		return nil, err
	}

	// Return the call expression node.
	return &ast.Call{
		Callee:    callee,
		Paren:     paren,     // This stores the right parenthesis token for error reporting.
		Arguments: arguments, // The list of parsed arguments.
	}, nil
}

func (p *Parser) primary() (ast.Expr, error) {
	if p.match(token.FALSE) {
		return &ast.Literal{Value: false, Line: p.previous().Line}, nil
	}
	if p.match(token.TRUE) {
		return &ast.Literal{Value: true, Line: p.previous().Line}, nil
	}
	if p.match(token.NIL) {
		return &ast.Literal{Value: nil, Line: p.previous().Line}, nil
	}

	if p.match(token.NUMBER, token.STRING) {
		return &ast.Literal{Value: p.previous().Literal, Line: p.previous().Line}, nil
	}

	if p.match(token.IDENTIFIER) {
		return &ast.Identifier{Name: p.previous(), Line: p.previous().Line}, nil
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

		return &ast.Grouping{Expression: expr, Line: p.previous().Line}, nil
	}

	// Parse array literals
	if p.match(token.LEFT_BRACKET) {
		return p.arrayLiteral()
	}

	// Parse object literals
	if p.match(token.LEFT_BRACE) {
		return p.objectLiteral()
	}

	return nil, p.error(p.peek(), "Unexpected token. Expect expression.")
}

func (p *Parser) objectLiteral() (ast.Expr, error) {
	properties := make(map[string]ast.Expr)

	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		propName, err := p.consume(token.IDENTIFIER, "Expect property name. Must be a string.")
		if err != nil {
			return nil, err
		}

		// Expect a colon `:` after the property name
		_, err = p.consume(token.COLON, "Expect ':' after property name.")
		if err != nil {
			return nil, err
		}

		// Parse the property value (expression)
		propValue, err := p.expression()
		if err != nil {
			return nil, err
		}

		// fmt.Printf("%#v ---- %#v\n", propName, propValue)
		// Store the property in the map
		properties[propName.Lexeme] = propValue

		// If there's no comma, break out of the loop
		if !p.match(token.COMMA) {
			break
		}
	}

	// Expect the closing right brace `}`
	_, err := p.consume(token.RIGHT_BRACE, "Expect '}' after object literal.")
	if err != nil {
		return nil, err
	}
	return &ast.ObjectLiteral{Properties: properties}, nil
}

// New function to handle array literals
func (p *Parser) arrayLiteral() (ast.Expr, error) {
	elements := []ast.Expr{}

	if !p.check(token.RIGHT_BRACKET) { // If the array is not empty
		for {
			element, err := p.expression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, element)

			// Check if there are more elements
			if !p.match(token.COMMA) {
				break
			}
		}
	}

	_, err := p.consume(token.RIGHT_BRACKET, "Expect ']' after array elements.")
	if err != nil {
		return nil, err
	}

	return &ast.ArrayLiteral{Elements: elements}, nil
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
