package interpreter

import (
	"fmt"

	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/token"
	"github.com/ah-naf/crafting-interpreter/utils"
)

func Eval(expr ast.Expr) interface{} {
	switch e := expr.(type) {
	case *ast.Literal:
		return e.Value

	case *ast.Grouping:
		return Eval(e.Expression)

	case *ast.Unary:
		right := Eval(e.Right)
		if utils.HadRuntimeError {
			return nil
		}
		return evaluateUnary(e.Operator, right)

	case *ast.Binary:
		left := Eval(e.Left)
		if utils.HadRuntimeError {
			return nil
		}
		right := Eval(e.Right)
		if utils.HadRuntimeError {
			return nil
		}
		return evaluateBinary(left, e.Operator, right)

	default:
		utils.RuntimeError(token.Token{Line: 0}, "Unknown expression type.")
		return nil
	}
}

func evaluateBinary(left interface{}, operator token.Token, right interface{}) interface{} {
	switch operator.Type {
	case token.PLUS:
		// Handle number addition and string concatenation
		switch l := left.(type) {
		case float64:
			switch r := right.(type) {
			case float64:
				// Number + Number
				return l + r
			case string:
				// Number + String -> Convert number to string and concatenate
				return fmt.Sprintf("%v", l) + r
			default:
				utils.RuntimeError(operator, "Right operand must be a number or string.")
				return nil
			}
		case string:
			switch r := right.(type) {
			case float64:
				// String + Number -> Convert number to string and concatenate
				return l + fmt.Sprintf("%v", r)
			case string:
				// String + String
				return l + r
			default:
				utils.RuntimeError(operator, "Right operand must be a number or string.")
				return nil
			}
		default:
			utils.RuntimeError(operator, "Left operand must be a number or string.")
			return nil
		}

	case token.MINUS:
		l, ok := left.(float64)
		if !ok {
			utils.RuntimeError(operator, "Left operand must be a number.")
			return nil
		}
		r, ok := right.(float64)
		if !ok {
			utils.RuntimeError(operator, "Right operand must be a number.")
			return nil
		}
		return l - r

	case token.STAR:
		l, ok := left.(float64)
		if !ok {
			utils.RuntimeError(operator, "Left operand must be a number.")
			return nil
		}
		r, ok := right.(float64)
		if !ok {
			utils.RuntimeError(operator, "Right operand must be a number.")
			return nil
		}
		return l * r

	case token.SLASH:
		l, ok := left.(float64)
		if !ok {
			utils.RuntimeError(operator, "Left operand must be a number.")
			return nil
		}
		r, ok := right.(float64)
		if !ok {
			utils.RuntimeError(operator, "Right operand must be a number.")
			return nil
		}
		if r == 0 {
			utils.RuntimeError(operator, "Divison by Zero")
			return nil
		}
		return l / r

	case token.EQUAL_EQUAL:
		return isEqual(left, right)

	case token.BANG_EQUAL:
		return !isEqual(left, right)

	case token.GREATER:
		l, ok := left.(float64)
		if !ok {
			utils.RuntimeError(operator, "Left operand must be a number.")
			return nil
		}
		r, ok := right.(float64)
		if !ok {
			utils.RuntimeError(operator, "Right operand must be a number.")
			return nil
		}
		return l > r

	case token.GREATER_EQUAL:
		l, ok := left.(float64)
		if !ok {
			utils.RuntimeError(operator, "Left operand must be a number.")
			return nil
		}
		r, ok := right.(float64)
		if !ok {
			utils.RuntimeError(operator, "Right operand must be a number.")
			return nil
		}
		return l >= r

	case token.LESS:
		l, ok := left.(float64)
		if !ok {
			utils.RuntimeError(operator, "Left operand must be a number.")
			return nil
		}
		r, ok := right.(float64)
		if !ok {
			utils.RuntimeError(operator, "Right operand must be a number.")
			return nil
		}
		return l < r

	case token.LESS_EQUAL:
		l, ok := left.(float64)
		if !ok {
			utils.RuntimeError(operator, "Left operand must be a number.")
			return nil
		}
		r, ok := right.(float64)
		if !ok {
			utils.RuntimeError(operator, "Right operand must be a number.")
			return nil
		}
		return l <= r

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
