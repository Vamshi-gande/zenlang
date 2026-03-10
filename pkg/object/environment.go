package object

// Environment is a runtime symbol table — it maps variable names to their
// current Object values. Every scope in Zen (global, function call, block)
// has its own Environment.
//
// The outer pointer implements lexical scoping: when a name is not found in
// the current store, the lookup walks up through outer environments.
// This is what makes local variables, function parameters, and closures work.
type Environment struct {
	store map[string]Object
	outer *Environment
}

// NewEnvironment creates a top-level environment with no enclosing scope.
// Used once for the global scope when the evaluator starts.
func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Object)}
}

// NewEnclosedEnvironment creates a new environment whose outer scope is the
// given environment. Called every time a function is invoked — the function
// body executes in this new scope, with the defining scope as its outer.
//
// This is the mechanism behind closures: the enclosed environment keeps a
// reference to the outer environment alive even after the outer function returns.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get looks up a variable name. It searches the current store first.
// If not found and an outer environment exists, the lookup is delegated upward.
// Returns the Object and true on success, nil and false if the name is not
// defined in any reachable scope.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set stores a name-to-value binding in the current environment's store.
// Always writes to the current scope — never walks up to outer scopes.
// Returns the stored object for convenience so callers can do:
//
//	return env.Set(name, value)
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
