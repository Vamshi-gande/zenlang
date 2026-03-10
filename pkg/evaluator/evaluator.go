package evaluator

import (
	"fmt"

	"github.com/Vamshi-gande/zenlang/pkg/ast"
	"github.com/Vamshi-gande/zenlang/pkg/object"
)

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

// Eval is the central dispatch function. It receives any AST node and the
// current environment, and returns the Object that the node evaluates to.
//
// Every evaluation path in the interpreter starts here. Nodes that contain
// child nodes (blocks, infix expressions, call expressions, etc.) call Eval
// recursively on their children.
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// -----------------------------------------------------------------------
	// Program — root of the AST
	// -----------------------------------------------------------------------
	case *ast.Program:
		return evalProgram(node, env)

	// -----------------------------------------------------------------------
	// Statements
	// -----------------------------------------------------------------------
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ReturnStatement:
		return evalReturnStatement(node, env)

	case *ast.LetStatement:
		return evalLetStatement(node, env)

	case *ast.WhileStatement:
		return evalWhileStatement(node, env)

	// -----------------------------------------------------------------------
	// Literal expressions — values that evaluate to themselves
	// -----------------------------------------------------------------------
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.NullLiteral:
		return object.NULL

	// -----------------------------------------------------------------------
	// Identifier — look up a name in the environment
	// -----------------------------------------------------------------------
	case *ast.Identifier:
		return evalIdentifier(node, env)

	// -----------------------------------------------------------------------
	// Operator expressions
	// -----------------------------------------------------------------------
	case *ast.PrefixExpression:
		return evalPrefixExpression(node, env)

	case *ast.InfixExpression:
		return evalInfixExpression(node, env)

	// -----------------------------------------------------------------------
	// Control flow
	// -----------------------------------------------------------------------
	case *ast.IfExpression:
		return evalIfExpression(node, env)

	// -----------------------------------------------------------------------
	// Functions
	// -----------------------------------------------------------------------
	case *ast.FunctionLiteral:
		return &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}

	case *ast.CallExpression:
		return evalCallExpression(node, env)

	// -----------------------------------------------------------------------
	// Data structures
	// -----------------------------------------------------------------------
	case *ast.ArrayLiteral:
		return evalArrayLiteral(node, env)

	case *ast.IndexExpression:
		return evalIndexExpression(node, env)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Program and block evaluation
// ---------------------------------------------------------------------------

// evalProgram evaluates each statement in sequence and returns the value of
// the last one. It is the only place that UNWRAPS a ReturnValue — a return
// at the top level should produce its inner value, not the wrapper.
// Errors also stop execution and are returned immediately.
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range program.Statements {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value // unwrap here — not inside function bodies
		case *object.Error:
			return result
		}
	}
	return result
}

// evalBlockStatement evaluates each statement in a block and returns the
// value of the last one. Unlike evalProgram it does NOT unwrap ReturnValues —
// it passes them upward so the enclosing function call can catch them.
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range block.Statements {
		result = Eval(stmt, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result // pass upward without unwrapping
			}
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Statement evaluation
// ---------------------------------------------------------------------------

// evalLetStatement evaluates the right-hand-side expression and binds the
// result to the variable name in the current environment.
// Returns nil (let statements have no value themselves).
func evalLetStatement(node *ast.LetStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}
	env.Set(node.Name.Value, val)
	return nil
}

// evalReturnStatement evaluates the return expression and wraps it in a
// ReturnValue object. The wrapper signals evalBlockStatement to stop executing
// further statements and bubble the value upward to the function call boundary.
func evalReturnStatement(node *ast.ReturnStatement, env *object.Environment) object.Object {
	val := Eval(node.ReturnValue, env)
	if isError(val) {
		return val
	}
	return &object.ReturnValue{Value: val}
}

// evalWhileStatement executes the body repeatedly as long as the condition
// is truthy. Each iteration re-evaluates the condition from scratch.
// A return or error inside the body terminates the loop and propagates upward.
func evalWhileStatement(node *ast.WhileStatement, env *object.Environment) object.Object {
	for {
		condition := Eval(node.Condition, env)
		if isError(condition) {
			return condition
		}
		if !isTruthy(condition) {
			break
		}
		result := Eval(node.Body, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return object.NULL
}

// ---------------------------------------------------------------------------
// Identifier lookup
// ---------------------------------------------------------------------------

// evalIdentifier resolves a variable name. It first checks the current
// environment chain, then falls back to the builtins map.
// An undefined name is a runtime error.
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: %s", node.Value)
}

// ---------------------------------------------------------------------------
// Prefix expressions
// ---------------------------------------------------------------------------

// evalPrefixExpression evaluates the right operand, then applies the
// prefix operator. Propagates errors from the right operand.
func evalPrefixExpression(node *ast.PrefixExpression, env *object.Environment) object.Object {
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}
	switch node.Operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalMinusPrefixOperator(right)
	case "++":
		return evalIncrementPrefix(node, right, env)
	case "--":
		return evalDecrementPrefix(node, right, env)
	default:
		return newError("unknown operator: %s%s", node.Operator, right.Type())
	}
}

// evalBangOperator applies logical NOT.
// !false → true, !null → true, !true → false, !<anything else> → false
func evalBangOperator(right object.Object) object.Object {
	switch right {
	case object.TRUE:
		return object.FALSE
	case object.FALSE:
		return object.TRUE
	case object.NULL:
		return object.TRUE
	default:
		return object.FALSE
	}
}

// evalMinusPrefixOperator negates an integer.
// Only valid on INTEGER — any other type is a type error.
func evalMinusPrefixOperator(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// evalIncrementPrefix handles ++x — increments the variable in the
// environment and returns the new value.
func evalIncrementPrefix(node *ast.PrefixExpression, right object.Object, env *object.Environment) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("operator ++ not supported for %s", right.Type())
	}
	ident, ok := node.Right.(*ast.Identifier)
	if !ok {
		return newError("operator ++ requires an identifier")
	}
	newVal := &object.Integer{Value: right.(*object.Integer).Value + 1}
	env.Set(ident.Value, newVal)
	return newVal
}

// evalDecrementPrefix handles --x — decrements the variable in the
// environment and returns the new value.
func evalDecrementPrefix(node *ast.PrefixExpression, right object.Object, env *object.Environment) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("operator -- not supported for %s", right.Type())
	}
	ident, ok := node.Right.(*ast.Identifier)
	if !ok {
		return newError("operator -- requires an identifier")
	}
	newVal := &object.Integer{Value: right.(*object.Integer).Value - 1}
	env.Set(ident.Value, newVal)
	return newVal
}

// ---------------------------------------------------------------------------
// Infix expressions
// ---------------------------------------------------------------------------

// evalInfixExpression evaluates both operands, then dispatches based on their
// types and the operator. Both operands are evaluated before any type checking,
// so errors in either propagate immediately.
func evalInfixExpression(node *ast.InfixExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	switch {
	case node.Operator == "=":
		return evalAssignment(node, right, env)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(node.Operator, left, right, node, env)
	case left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(node.Operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(node.Operator, left, right)
	case node.Operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case node.Operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case node.Operator == "&&":
		return nativeBoolToBooleanObject(isTruthy(left) && isTruthy(right))
	case node.Operator == "||":
		return nativeBoolToBooleanObject(isTruthy(left) || isTruthy(right))
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
	}
}

// evalAssignment handles bare assignment: x = expr.
// The left side must be an Identifier — anything else is a runtime error.
// Uses env.Update which walks the scope chain to mutate the existing binding
// in place — this is what makes closure mutation work.
func evalAssignment(node *ast.InfixExpression, val object.Object, env *object.Environment) object.Object {
	ident, ok := node.Left.(*ast.Identifier)
	if !ok {
		return newError("assignment target must be an identifier, got %T", node.Left)
	}
	env.Update(ident.Value, val)
	return val
}

// evalIntegerInfixExpression handles all infix operations where both operands
// are integers. Compound assignment operators (+=, -=, *=, /=) mutate the
// left-hand variable in the environment and return the new value.
func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
	node *ast.InfixExpression,
	env *object.Environment,
) object.Object {
	lv := left.(*object.Integer).Value
	rv := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: lv + rv}
	case "-":
		return &object.Integer{Value: lv - rv}
	case "*":
		return &object.Integer{Value: lv * rv}
	case "/":
		if rv == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: lv / rv}
	case "<":
		return nativeBoolToBooleanObject(lv < rv)
	case ">":
		return nativeBoolToBooleanObject(lv > rv)
	case "<=":
		return nativeBoolToBooleanObject(lv <= rv)
	case ">=":
		return nativeBoolToBooleanObject(lv >= rv)
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)

	// Compound assignments — require the left side to be an identifier
	case "+=", "-=", "*=", "/=":
		ident, ok := node.Left.(*ast.Identifier)
		if !ok {
			return newError("compound assignment requires an identifier on the left side")
		}
		var newVal int64
		switch operator {
		case "+=":
			newVal = lv + rv
		case "-=":
			newVal = lv - rv
		case "*=":
			newVal = lv * rv
		case "/=":
			if rv == 0 {
				return newError("division by zero")
			}
			newVal = lv / rv
		}
		result := &object.Integer{Value: newVal}
		env.Set(ident.Value, result)
		return result

	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalFloatInfixExpression handles arithmetic and comparisons when at least
// one operand is a float. The integer operand is promoted to float64.
func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	var lv, rv float64
	switch l := left.(type) {
	case *object.Float:
		lv = l.Value
	case *object.Integer:
		lv = float64(l.Value)
	default:
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	}
	switch r := right.(type) {
	case *object.Float:
		rv = r.Value
	case *object.Integer:
		rv = float64(r.Value)
	default:
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	}

	switch operator {
	case "+":
		return &object.Float{Value: lv + rv}
	case "-":
		return &object.Float{Value: lv - rv}
	case "*":
		return &object.Float{Value: lv * rv}
	case "/":
		if rv == 0 {
			return newError("division by zero")
		}
		return &object.Float{Value: lv / rv}
	case "<":
		return nativeBoolToBooleanObject(lv < rv)
	case ">":
		return nativeBoolToBooleanObject(lv > rv)
	case "<=":
		return nativeBoolToBooleanObject(lv <= rv)
	case ">=":
		return nativeBoolToBooleanObject(lv >= rv)
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalStringInfixExpression handles string operations. Only + (concatenation)
// and == / != are supported. Any other operator is a type error.
func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	lv := left.(*object.String).Value
	rv := right.(*object.String).Value
	switch operator {
	case "+":
		return &object.String{Value: lv + rv}
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// ---------------------------------------------------------------------------
// Control flow
// ---------------------------------------------------------------------------

// evalIfExpression evaluates the condition and then evaluates either the
// consequence or the alternative block. If there is no alternative and the
// condition is falsy, returns NULL.
func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		return Eval(node.Alternative, env)
	}
	return object.NULL
}

// isTruthy defines Zen's truthiness rules.
// NULL and FALSE are falsy. Everything else — including 0, empty string,
// empty array — is truthy.
func isTruthy(obj object.Object) bool {
	switch obj {
	case object.NULL:
		return false
	case object.TRUE:
		return true
	case object.FALSE:
		return false
	default:
		return true
	}
}

// ---------------------------------------------------------------------------
// Functions
// ---------------------------------------------------------------------------

// evalCallExpression evaluates the callee expression, evaluates all argument
// expressions left-to-right, then applies the function to the arguments.
func evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {
	function := Eval(node.Function, env)
	if isError(function) {
		return function
	}
	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	return applyFunction(function, args)
}

// evalExpressions evaluates a slice of expressions in order and returns their
// objects. If any expression produces an error, evaluation stops immediately
// and a single-element slice containing the error is returned.
func evalExpressions(exprs []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exprs {
		val := Eval(e, env)
		if isError(val) {
			return []object.Object{val}
		}
		result = append(result, val)
	}
	return result
}

// applyFunction dispatches a call to either a user-defined Function or a
// built-in Builtin. Anything else is a "not a function" runtime error.
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

// extendFunctionEnv creates a new enclosed environment using the function's
// captured defining environment as the outer scope, then binds each parameter
// name to its corresponding argument value.
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}
	return env
}

// unwrapReturnValue strips the ReturnValue wrapper if present.
// This is called at the function call boundary — the wrapper has done its job
// of bubbling up through nested blocks, and now the plain value should be
// returned to the caller.
func unwrapReturnValue(obj object.Object) object.Object {
	if rv, ok := obj.(*object.ReturnValue); ok {
		return rv.Value
	}
	return obj
}

// ---------------------------------------------------------------------------
// Data structures
// ---------------------------------------------------------------------------

// evalArrayLiteral evaluates all element expressions and wraps them in an
// Array object.
func evalArrayLiteral(node *ast.ArrayLiteral, env *object.Environment) object.Object {
	elements := evalExpressions(node.Elements, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}
	return &object.Array{Elements: elements}
}

// evalIndexExpression evaluates the object being indexed and the index itself,
// then dispatches to the appropriate index handler based on the object type.
func evalIndexExpression(node *ast.IndexExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

// evalArrayIndexExpression retrieves the element at the given integer index.
// Out-of-bounds and negative indices return NULL.
func evalArrayIndexExpression(array, index object.Object) object.Object {
	arr := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arr.Elements) - 1)
	if idx < 0 || idx > max {
		return object.NULL
	}
	return arr.Elements[idx]
}

// evalHashIndexExpression looks up a key in a Hash. The key must implement
// Hashable — if it does not, a type error is returned.
// A missing key returns NULL rather than an error.
func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return object.NULL
	}
	return pair.Value
}

// evalHashLiteral evaluates each key and value expression pair, verifies the
// key is Hashable, and builds the Pairs map.
func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}
		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// nativeBoolToBooleanObject returns the TRUE or FALSE singleton.
// Never allocates a new Boolean — always reuses the singletons.
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return object.TRUE
	}
	return object.FALSE
}

// newError creates an Error object with a formatted message.
// Used throughout the evaluator wherever a runtime error must be produced.
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// isError checks whether an object is an Error. Called after almost every
// Eval call to ensure errors propagate immediately without further evaluation.
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
