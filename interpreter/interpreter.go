package interpreter

import (
	"fmt"
	"math"

	"github.com/ah-naf/crafting-interpreter/ast"
	"github.com/ah-naf/crafting-interpreter/environment"
	"github.com/ah-naf/crafting-interpreter/token"
	"github.com/ah-naf/crafting-interpreter/utils"
)

// Interpreter struct represents the execution context for evaluating expressions and statements.
type Interpreter struct {
	env     *environment.Environment
	globals *environment.Environment
}

type ControlFlowSignal struct {
	Type       int
	LineNumber int
}

// NewInterpreter creates a new instance of the Interpreter with the given environment.
func NewInterpreter() *Interpreter {
	// Define the global environment and set up the clock function first
	globals := environment.NewEnvironment()

	globals.Define("clock", NativeClockFn{})

	// Then, create the Interpreter instance with the global environment
	i := &Interpreter{
		env:     globals, // The interpreter starts with the global environment
		globals: globals, // Store the reference to the global environment
	}

	return i
}

const (
	ControlFlowNone int = iota
	ControlFlowBreak
	ControlFlowContinue
)

func (i *Interpreter) Interpret(statements []ast.Stmt, isRepl bool) []interface{} {
	var results []interface{}
	env := environment.NewEnvironment()

	for _, statement := range statements {
		// fmt.Printf("%#v\n", statement)
		result, signal := i.eval(statement, env, isRepl)
		if signal.Type == ControlFlowBreak {
			utils.RuntimeError(token.Token{Line: signal.LineNumber}, "Unexpected 'break' outside of loop.")
			return nil
		} else if signal.Type == ControlFlowContinue {
			utils.RuntimeError(token.Token{Line: signal.LineNumber}, "Unexpected 'continue' outside of loop.")
			return nil
		}
		// fmt.Printf("%#v\n", result)
		if utils.HadRuntimeError {
			return nil // Stop execution if a runtime error occurred during evaluation
		}
		results = append(results, result)
	}

	return results
}

func (i *Interpreter) eval(expr ast.Expr, env *environment.Environment, isRepl bool) (interface{}, *ControlFlowSignal) {
	switch e := expr.(type) {
	case *ast.FunctionStmt:
		function := NewFunction(e)
		// fmt.Printf("%#v %#v\n",e.Name.Lexeme, function)
		env.Define(e.Name.Lexeme, function)
		return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.Call:
		// Step 1: Evaluate the callee (the thing being called)
		
		callee, signal := i.eval(e.Callee, env, isRepl)

		
		if signal.Type != ControlFlowNone {
			return nil, signal
		}

		// Ensure the callee is a callable function
		function, ok := callee.(Callable)
		if len(e.Arguments) != function.Arity() {
			utils.RuntimeError(e.Paren, fmt.Sprintf("Expected %d arguments but %d.", function.Arity(), len(e.Arguments)))
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
		}

		if !ok {
			utils.RuntimeError(e.Paren, "Can only call functions or classes.")
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
		}

		// Step 2: Evaluate each argument and collect them in a list
		var arguments []interface{}
		for _, arg := range e.Arguments {
			argValue, signal := i.eval(arg, env, isRepl)
			if signal.Type != ControlFlowNone {
				return nil, signal
			}
			arguments = append(arguments, argValue)
		}

		// Step 3: Call the function and return its result
		result, err := function.Call(i, arguments)
		if err != nil {
			utils.RuntimeError(e.Paren, "Function call failed: "+err.Error())
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
		}

		return result, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.PrintStatement:
		value, signal := i.eval(e.Expression, env, isRepl)
		if signal.Type != ControlFlowNone {
			return value, signal
		}
		if utils.HadRuntimeError {
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0} // Stop execution if a runtime error occurred during evaluation
		}
		fmt.Println(stringify(value))
		return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.ExpressionStatement:
		value, signal := i.eval(e.Expression, env, isRepl)

		if signal.Type != ControlFlowNone {
			return nil, signal
		}
		if isRepl && !utils.HadRuntimeError {
			fmt.Println(stringify(value))
		}
		return value, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.Literal:
		return e.Value, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.Grouping:
		return i.eval(e.Expression, env, isRepl)

	case *ast.Unary:
		right, signal := i.eval(e.Right, env, isRepl)
		if signal.Type != ControlFlowNone {
			return nil, signal
		}
		if utils.HadRuntimeError {
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
		}
		return evaluateUnary(e.Operator, right), &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.Binary:
		left, signal := i.eval(e.Left, env, isRepl)
		if signal.Type != ControlFlowNone {
			return nil, signal
		}
		if utils.HadRuntimeError {
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
		}
		right, signal := i.eval(e.Right, env, isRepl)
		if signal.Type != ControlFlowNone {
			return nil, signal
		}
		if utils.HadRuntimeError {
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
		}
		return evaluateBinary(left, e.Operator, right), &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.VarStmt:
		var value interface{}
		if e.Initializer != nil {
			v, signal := i.eval(e.Initializer, env, isRepl)
			if signal.Type != ControlFlowNone {
				return nil, signal
			}
			if utils.HadRuntimeError {
				return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
			}
			value = v
		}
		_, err := env.Get(e.Name.Lexeme)
		if err != nil {
			env.Define(e.Name.Lexeme, value)
		} else {
			utils.RuntimeError(token.Token{Line: e.Line}, "Cannot redeclare variable "+e.Name.Lexeme+".")
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
		}
		return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.VarListStmt:
		for _, decl := range e.Declarations {
			_, signal := i.eval(&decl, env, isRepl)
			if signal.Type != ControlFlowNone {
				return nil, signal
			}
			if utils.HadRuntimeError {
				return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
			}
		}
		return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.AssignmentStmt:
		val, signal := i.eval(e.Value, env, isRepl)
		if signal.Type != ControlFlowNone {
			return nil, signal
		}
		if utils.HadRuntimeError {
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
		}
		env.Assign(e.Name, val)
		return val, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.Identifier:
		val, err := env.Get(e.Name.Lexeme)
		if err != nil {
			utils.RuntimeError(token.Token{Line: e.Line}, "Variable "+e.Name.Lexeme+" is not defined.")
			return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
		}
		return val, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.BlockStmt:
		newEnv := environment.NewEnvironmentWithParent(env)
		for _, statement := range e.Block {
			_, signal := i.eval(statement, newEnv, isRepl)
			if signal.Type != ControlFlowNone {
				return nil, signal
			}
			if utils.HadRuntimeError {
				return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
			}
		}
		return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.IfStmt:
		cc, signal := i.eval(e.Condition, env, isRepl)
		if signal.Type != ControlFlowNone {
			return nil, signal
		}
		if isTruthy(cc) {
			_, signal := i.eval(e.ThenBranch, env, isRepl)
			if signal.Type != ControlFlowNone {
				return nil, signal
			}
		} else if e.ElseBranch != nil {
			_, signal := i.eval(e.ElseBranch, env, isRepl)
			if signal.Type != ControlFlowNone {
				return nil, signal
			}
		}
		return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.Logical:
		left, signal := i.eval(e.Left, env, isRepl)
		if signal.Type != ControlFlowNone {
			return nil, signal
		}
		// fmt.Printf("%v %v %v\n", left, e.Operator.Type, token.OR)
		if e.Operator.Type == token.LOGICAL_OR {
			if isTruthy(left) {
				return left, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
			}
		} else {
			if !isTruthy(left) {
				return left, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
			}
		}
		return i.eval(e.Right, env, isRepl)

	case *ast.While:
		for {
			condVal, signal := i.eval(e.Condition, env, isRepl)
			if signal.Type != ControlFlowNone {
				return nil, signal // Propagate signal upwards
			}
			if !isTruthy(condVal) {
				break
			}

			_, signal = i.eval(e.Body, env, isRepl)
			if signal.Type == ControlFlowBreak {
				break // Exit the loop
			}
		}
		return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.ForStmt:
		// Execute the initializer
		if e.Initializer != nil {
			_, signal := i.eval(e.Initializer, env, isRepl)
			if signal.Type != ControlFlowNone {
				return nil, signal
			}
		}

		for {
			// Check the condition
			if e.Condition != nil {
				condVal, signal := i.eval(e.Condition, env, isRepl)
				if signal.Type != ControlFlowNone {
					return nil, signal
				}
				if !isTruthy(condVal) {
					break
				}
			}

			// Execute the body
			_, signal := i.eval(e.Body, env, isRepl)
			if signal.Type == ControlFlowBreak {
				break
			}
			if signal.Type == ControlFlowContinue {
				// Skip to the increment
			} else if signal.Type != ControlFlowNone {
				return nil, signal
			}

			// Execute the increment
			if e.Increment != nil {
				_, signal := i.eval(e.Increment, env, isRepl)
				if signal.Type != ControlFlowNone {
					return nil, signal
				}
			}
		}
		return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}

	case *ast.BreakStmt:
		return nil, &ControlFlowSignal{Type: ControlFlowBreak, LineNumber: e.Line}

	case *ast.ContinueStmt:
		return nil, &ControlFlowSignal{Type: ControlFlowContinue, LineNumber: e.Line}

	default:
		lineNumber := getLineNumber(expr)
		utils.RuntimeError(token.Token{Line: lineNumber}, "Unknown expression type.")
		return nil, &ControlFlowSignal{Type: ControlFlowNone, LineNumber: 0}
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

	case token.AND, token.OR, token.XOR, token.LEFT_SHIFT, token.RIGHT_SHIFT:
		return handleBitwise(left, right, operator)

	case token.POWER:
		leftFloat, err := toNumber(left)
		if err != nil {
			utils.RuntimeError(operator, "Left operand must be a number.")
			return nil
		}
		rightFloat, err := toNumber(right)
		if err != nil {
			utils.RuntimeError(operator, "Right operand must be a number.")
			return nil
		}
		return math.Pow(leftFloat, rightFloat)

	case token.MODULO:
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
		if rightNum == 0 {
			utils.RuntimeError(operator, "Division by zero.")
			return nil
		}
		return math.Mod(leftNum, rightNum)

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
	if num, ok := value.(float64); ok {
		// 0.0 should be false, any non-zero number should be true
		return num != 0.0
	}
	if str, ok := value.(string); ok {
		// An empty string should be false, non-empty string should be true
		return str != ""
	}
	return true // Everything else is considered true
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
	case *ast.BreakStmt:
		return e.Line
	case *ast.ContinueStmt:
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
