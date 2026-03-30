package main

import (
	"fmt"
	"os"

	"github.com/Vamshi-gande/zenlang/pkg/evaluator"
	"github.com/Vamshi-gande/zenlang/pkg/lexer"
	"github.com/Vamshi-gande/zenlang/pkg/object"
	"github.com/Vamshi-gande/zenlang/pkg/parser"
	"github.com/Vamshi-gande/zenlang/pkg/repl"
)

func main() {
	if len(os.Args) > 1 {
		runFile(os.Args[1])
	} else {
		printWelcome()
		repl.New(os.Stdin, os.Stdout).Start()
	}
}

// printWelcome prints the greeting shown when the REPL starts interactively.
func printWelcome() {
	fmt.Println("Welcome to Zen v0.1")
	fmt.Println("Type 'exit' or 'quit' to leave")
	fmt.Println()
}

// runFile reads a .zen source file, runs it through the full pipeline once,
// and exits. Parse errors and runtime errors are written to stderr with a
// non-zero exit code so that scripts can be used in shell pipelines.
func runFile(filename string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file '%s': %s\n", filename, err)
		os.Exit(1)
	}

	l := lexer.NewLexer(string(source))
	p := parser.NewParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "parse errors:")
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "\t%s\n", e.Error())
		}
		os.Exit(1)
	}

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)

	if result != nil && result.Type() == object.ERROR_OBJ {
		fmt.Fprintln(os.Stderr, result.Inspect())
		os.Exit(1)
	}
}
