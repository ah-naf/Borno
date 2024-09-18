package interpreter

import (
	"fmt"
	"math"

	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/environment"
	"github.com/ah-naf/crafting-interpreter/token"
	"github.com/ah-naf/crafting-interpreter/utils"
)

func Interpret(statements []ast.Stmt, isRepl bool) []interface{} {
	var results []interface{}
	env := environment.NewEnvironment()

	for _, statement := range statements {
		// fmt.Printf("%#v\n", statement)
		result := eval(statement, env, isRepl)
		// fmt.Printf("%#v\n", result)
		if utils.HadRuntimeError {
			return nil // Stop execution if a runtime error occurred during evaluation
		}
		results = append(results, result)
	}

	return results
}

func eval(expr ast.Expr, env *environment.Environment, isRepl bool) interface{} {
	switch e := expr.(type) {
	case *ast.PrintStatement:
		value := eval(e.Expression, env, isRepl)
		if utils.HadRuntimeError {
			return nil // Stop execution if a runtime error occurred during evaluation
		}
		fmt.Println(stringify(value))
		return nil

	case *ast.ExpressionStatement:
		value := eval(e.Expression, env, isRepl)
		if isRepl && !utils.HadRuntimeError {
			fmt.Println(stringify(value))
		}
		return value

	case *ast.Literal:
		return e.Value

	case *ast.Grouping:
		return eval(e.Expression, env, isRepl)

	case *ast.Unary:
		right := eval(e.Right, env, isRepl)
		if utils.HadRuntimeError {
			return nil
		}
		return evaluateUnary(e.Operator, right)

	case *ast.Binary:
		left := eval(e.Left, env, isRepl)
		if utils.HadRuntimeError {
			return nil
		}
		right := eval(e.Right, env, isRepl)
		if utils.HadRuntimeError {
			return nil
		}
		return evaluateBinary(left, e.Operator, right)

	case *ast.VarStmt:
		var value interface{}
		if e.Initializer != nil {
			value = eval(e.Initializer, env, isRepl)
			if utils.HadRuntimeError {
				return nil
			}
		}
		_, err := env.Get(e.Name.Lexeme)
		if err != nil {
			env.Define(e.Name.Lexeme, value)
		} else {
			utils.RuntimeError(token.Token{Line: e.Line}, "Cannot redeclare variable "+e.Name.Lexeme+".")
			return nil
		}
		return nil

	case *ast.VarListStmt:
		for _, decl := range e.Declarations {
			eval(&decl, env, isRepl)
			if utils.HadRuntimeError {
				return nil
			}
		}
		return nil

	case *ast.AssignmentStmt:
		val := eval(e.Value, env, isRepl)
		if utils.HadRuntimeError {
			return nil
		}
		env.Assign(e.Name, val)
		return val

	case *ast.Identifier:
		val, err := env.Get(e.Name.Lexeme)
		if err != nil {
			utils.RuntimeError(token.Token{Line: e.Line}, "Variable "+e.Name.Lexeme+" is not defined.")
			return nil
		}
		return val

	case *ast.BlockStmt:
		newEnv := environment.NewEnvironmentWithParent(env)
		for _, statement := range e.Block {
			eval(statement, newEnv, isRepl)
			if utils.HadRuntimeError {
				return nil
			}
		}
		return nil

	default:
		lineNumber := getLineNumber(expr)
		utils.RuntimeError(token.Token{Line: lineNumber}, "Unknown expression type.")
		return nil
	}
}

func evaluateBinary(left interface{}, operator token.Token, right interface{}) interface{} {
	if utils.HadRuntimeError {
		return nil
	}

	switch operator.Type {
	case token.PLUS:
		return handleAddition(left, right, operator)

	case token.MINUS, token.STAR, token.SLASH:
		return handleArithmetic(left, right, operator)

	case token.EQUAL_EQUAL, token.BANG_EQUAL:
		return handleEquality(left, right, operator)

	case token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL:
		return handleComparison(left, right, operator)

	case token.AND, token.OR, token.XOR, token.LEFT_SHIFT, token.RIGHT_SHIFT, token.POWER:
		return handleBitwise(left, right, operator)

	default:
		utils.RuntimeError(operator, "Unknown binary operator: "+operator.Lexeme)
		return nil
	}
}

func evaluateUnary(operator token.Token, right interface{}) interface{} {
	if utils.HadRuntimeError {
		return nil
	}

	switch operator.Type {
	case token.MINUS:
		value, err := toNumber(right)
		if err != nil {
			utils.RuntimeError(operator, err.Error())
			return nil
		}
		return -value

	case token.BANG:
		return !isTruthy(right)

	case token.NOT:
		value, err := toInt64(right)
		if err != nil {
			utils.RuntimeError(operator, err.Error())
			return nil
		}
		return ^value

	default:
		utils.RuntimeError(operator, "Unknown unary operator: "+operator.Lexeme)
		return nil
	}
}

// Helper functions to reduce code duplication

func handleAddition(left, right interface{}, operator token.Token) interface{} {
	// Handle number addition and string concatenation
	switch l := left.(type) {
	case int64, float64:
		leftNum, err := toNumber(left)
		if err != nil {
			utils.RuntimeError(operator, "Left operand must be a number.")
			return nil
		}
		rightNum, err := toNumber(right)
		if err == nil {
			return leftNum + rightNum
		}
		rightStr, ok := right.(string)
		if ok {
			return fmt.Sprintf("%v", leftNum) + rightStr
		}
	case string:
		rightStr, err := stringifyOperand(right)
		if err != nil {
			utils.RuntimeError(operator, "Right operand must be a string or number.")
			return nil
		}
		return l + rightStr
	}
	utils.RuntimeError(operator, "Operands must be numbers or strings.")
	return nil
}

func handleArithmetic(left, right interface{}, operator token.Token) interface{} {
	leftNum, err := toNumber(left)
	if err != nil {
		utils.RuntimeError(operator, "Left operand must be a number.")
		return nil
	}
	rightNum, err := toNumber(right)
	if err != nil {
		utils.RuntimeError(operator, "Right operand must be a number.")
		return nil
	}

	switch operator.Type {
	case token.MINUS:
		return leftNum - rightNum
	case token.STAR:
		return leftNum * rightNum
	case token.SLASH:
		if rightNum == 0 {
			utils.RuntimeError(operator, "Division by zero.")
			return nil
		}
		return leftNum / rightNum
	}
	return nil
}

func handleEquality(left, right interface{}, operator token.Token) interface{} {
	isEqual := isEqual(left, right)
	if operator.Type == token.BANG_EQUAL {
		return !isEqual
	}
	return isEqual
}

func handleComparison(left, right interface{}, operator token.Token) interface{} {
	leftNum, err := toNumber(left)
	if err != nil {
		utils.RuntimeError(operator, "Left operand must be a number.")
		return nil
	}
	rightNum, err := toNumber(right)
	if err != nil {
		utils.RuntimeError(operator, "Right operand must be a number.")
		return nil
	}

	switch operator.Type {
	case token.GREATER:
		return leftNum > rightNum
	case token.GREATER_EQUAL:
		return leftNum >= rightNum
	case token.LESS:
		return leftNum < rightNum
	case token.LESS_EQUAL:
		return leftNum <= rightNum
	}
	return nil
}

func handleBitwise(left, right interface{}, operator token.Token) interface{} {
	leftInt, err := toInt64(left)
	if err != nil {
		utils.RuntimeError(operator, "Left operand must be an integer.")
		return nil
	}
	rightInt, err := toInt64(right)
	if err != nil {
		utils.RuntimeError(operator, "Right operand must be an integer.")
		return nil
	}

	switch operator.Type {
	case token.AND:
		return leftInt & rightInt
	case token.OR:
		return leftInt | rightInt
	case token.XOR:
		return leftInt ^ rightInt
	case token.LEFT_SHIFT:
		return leftInt << rightInt
	case token.RIGHT_SHIFT:
		return leftInt >> rightInt
	case token.POWER:
		return int64(math.Pow(float64(leftInt), float64(rightInt)))
	}
	return nil
}

// Helper functions for type conversions

func toNumber(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("expected a number, got %T", value)
	}
}

func toInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case float64:
		if float64(int64(v)) == v {
			return int64(v), nil
		}
		return 0, fmt.Errorf("expected an integer, got float %v", v)
	default:
		return 0, fmt.Errorf("expected an integer, got %T", value)
	}
}

func stringifyOperand(value interface{}) (string, error) {
	switch v := value.(type) {
	case int64, float64, string:
		return fmt.Sprintf("%v", v), nil
	default:
		return "", fmt.Errorf("cannot stringify value of type %T", value)
	}
}

func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	return true
}

func isEqual(a, b interface{}) bool {
	return a == b
}

func getLineNumber(expr ast.Expr) int {
	switch e := expr.(type) {
	case *ast.Binary:
		return e.Line
	case *ast.Unary:
		return e.Line
	case *ast.Literal:
		return e.Line
	case *ast.Grouping:
		return e.Line
	case *ast.VarStmt:
		return e.Name.Line
	case *ast.Identifier:
		return e.Line
	// Add cases for other expression types if necessary
	default:
		return 0 // Return 0 if line number is not available
	}
}

func stringify(value interface{}) string {
	if value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", value)
}
