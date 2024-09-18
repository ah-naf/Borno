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
		result := eval(statement, env, isRepl)
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
		// Return value if you want to make print return the expression
		return nil

	case *ast.ExpressionStatement:
		value := eval(e.Expression, env, isRepl)
		if isRepl {
			fmt.Println(stringify(value))
		}
		return value

	case *ast.Literal:
		return e.Value

	case *ast.Grouping:
		return eval(e.Expression, env, isRepl)

	case *ast.Unary:
		right := eval(e.Right, env, isRepl)

		return evaluateUnary(e.Operator, right)

	case *ast.Binary:
		left := eval(e.Left, env, isRepl)

		right := eval(e.Right, env, isRepl)
		if utils.HadRuntimeError {
			return nil
		}
		return evaluateBinary(left, e.Operator, right)

	case *ast.VarStmt:
		var value interface{}

		if e.Initializer != nil {
			value = eval(e.Initializer, env, isRepl)
		}

		env.Define(e.Name.Lexeme, value)
		return nil

	case *ast.AssignmentStmt:
		val := eval(e.Value, env, isRepl)
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
		}
		return nil

	default:
		lineNumber := getLineNumber(expr)
		utils.RuntimeError(token.Token{Line: lineNumber}, "Unknown expression type.")
		return nil
	}
}

func evaluateBinary(left interface{}, operator token.Token, right interface{}) interface{} {
	// Helper function to determine if both operands are int64
	isInt64 := func(val interface{}) (int64, bool) {
		v, ok := val.(int64)
		if ok {
			return v, true
		}
		if f, ok := val.(float64); ok && float64(int64(f)) == f {
			return int64(f), true
		}
		return 0, false
	}

	switch operator.Type {
	case token.PLUS:
		// Handle number addition and string concatenation
		switch l := left.(type) {
		case int64:
			switch r := right.(type) {
			case int64:
				// int64 + int64
				return l + r
			case float64:
				// int64 + float64
				return float64(l) + r
			case string:
				// int64 + string -> Convert int64 to string and concatenate
				return fmt.Sprintf("%v", l) + r
			}
		case float64:
			switch r := right.(type) {
			case int64:
				// float64 + int64
				return l + float64(r)
			case float64:
				// float64 + float64
				return l + r
			case string:
				// float64 + string -> Convert float64 to string and concatenate
				return fmt.Sprintf("%v", l) + r
			}
		case string:
			switch r := right.(type) {
			case int64:
				// string + int64 -> Convert int64 to string and concatenate
				return l + fmt.Sprintf("%v", r)
			case float64:
				// string + float64 -> Convert float64 to string and concatenate
				return l + fmt.Sprintf("%v", r)
			case string:
				// string + string
				return l + r
			}
		}
		utils.RuntimeError(operator, "Operands must be numbers or strings.")
		return nil

	case token.MINUS:
		switch l := left.(type) {
		case int64:
			switch r := right.(type) {
			case int64:
				// int64 - int64
				return l - r
			case float64:
				// int64 - float64
				return float64(l) - r
			}
		case float64:
			switch r := right.(type) {
			case int64:
				// float64 - int64
				return l - float64(r)
			case float64:
				// float64 - float64
				return l - r
			}
		}
		utils.RuntimeError(operator, "Operands must be numbers.")
		return nil

	case token.STAR:
		switch l := left.(type) {
		case int64:
			switch r := right.(type) {
			case int64:
				// int64 * int64
				return l * r
			case float64:
				// int64 * float64
				return float64(l) * r
			}
		case float64:
			switch r := right.(type) {
			case int64:
				// float64 * int64
				return l * float64(r)
			case float64:
				// float64 * float64
				return l * r
			}
		}
		utils.RuntimeError(operator, "Operands must be numbers.")
		return nil

	case token.SLASH:
		switch l := left.(type) {
		case int64:
			switch r := right.(type) {
			case int64:
				if r == 0 {
					utils.RuntimeError(operator, "Division by zero.")
					return nil
				}
				return l / r
			case float64:
				if r == 0 {
					utils.RuntimeError(operator, "Division by zero.")
					return nil
				}
				return float64(l) / r
			}
		case float64:
			switch r := right.(type) {
			case int64:
				if r == 0 {
					utils.RuntimeError(operator, "Division by zero.")
					return nil
				}
				return l / float64(r)
			case float64:
				if r == 0 {
					utils.RuntimeError(operator, "Division by zero.")
					return nil
				}
				return l / r
			}
		}
		utils.RuntimeError(operator, "Operands must be numbers.")
		return nil

	case token.EQUAL_EQUAL:
		return isEqual(left, right)

	case token.BANG_EQUAL:
		return !isEqual(left, right)

	case token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL:
		switch l := left.(type) {
		case int64:
			switch r := right.(type) {
			case int64:
				switch operator.Type {
				case token.GREATER:
					return l > r
				case token.GREATER_EQUAL:
					return l >= r
				case token.LESS:
					return l < r
				case token.LESS_EQUAL:
					return l <= r
				}
			case float64:
				switch operator.Type {
				case token.GREATER:
					return float64(l) > r
				case token.GREATER_EQUAL:
					return float64(l) >= r
				case token.LESS:
					return float64(l) < r
				case token.LESS_EQUAL:
					return float64(l) <= r
				}
			}
		case float64:
			switch r := right.(type) {
			case int64:
				switch operator.Type {
				case token.GREATER:
					return l > float64(r)
				case token.GREATER_EQUAL:
					return l >= float64(r)
				case token.LESS:
					return l < float64(r)
				case token.LESS_EQUAL:
					return l <= float64(r)
				}
			case float64:
				switch operator.Type {
				case token.GREATER:
					return l > r
				case token.GREATER_EQUAL:
					return l >= r
				case token.LESS:
					return l < r
				case token.LESS_EQUAL:
					return l <= r
				}
			}
		}
		utils.RuntimeError(operator, "Operands must be numbers.")
		return nil

	case token.AND, token.OR, token.XOR, token.LEFT_SHIFT, token.RIGHT_SHIFT, token.POWER:
		// Ensure both operands are integers
		lInt, lIsInt := isInt64(left)
		rInt, rIsInt := isInt64(right)

		if !lIsInt || !rIsInt {
			utils.RuntimeError(operator, "Bitwise operators can only be applied to integers.")
			return nil
		}

		switch operator.Type {
		case token.AND:
			return lInt & rInt
		case token.OR:
			return lInt | rInt
		case token.XOR:
			return lInt ^ rInt
		case token.LEFT_SHIFT:
			return lInt << rInt
		case token.RIGHT_SHIFT:
			return lInt >> rInt
		case token.POWER:
			return int64(math.Pow(float64(lInt), float64(rInt)))
		default:
			utils.RuntimeError(operator, "Unknown binary operator: "+operator.Lexeme)
			return nil
		}

	default:
		utils.RuntimeError(operator, "Unknown binary operator: "+operator.Lexeme)
		return nil
	}
}

func evaluateUnary(operator token.Token, right interface{}) interface{} {
	switch operator.Type {
	case token.MINUS:
		value, ok := right.(float64)
		if !ok {
			utils.RuntimeError(operator, "Operand must be a number.")
			return nil
		}
		return -value

	case token.BANG:
		return !isTruthy(right)
	case token.NOT:
		rFloat, rok := right.(float64)

		if !rok {
			utils.RuntimeError(operator, "Operands must be numbers.")
			return nil
		}

		// Check if the operands are actually integers
		if float64(int64(rFloat)) != rFloat {
			utils.RuntimeError(operator, "Bitwise operators can only be applied to integers.")
			return nil
		}

		return ^int64(rFloat)
	default:
		utils.RuntimeError(operator, "Unknown unary operator: "+operator.Lexeme)
		return nil
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
	// fmt.Printf("%#v\n", expr)
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
