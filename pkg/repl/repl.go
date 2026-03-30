package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/Vamshi-gande/zenlang/pkg/evaluator"
	"github.com/Vamshi-gande/zenlang/pkg/lexer"
	"github.com/Vamshi-gande/zenlang/pkg/object"
	"github.com/Vamshi-gande/zenlang/pkg/parser"
)

// PROMPT is the string printed at the start of every input line.
const PROMPT = "zen> "

// REPL holds the input/output streams and the persistent environment.
// Using io.Reader and io.Writer (rather than hardcoded os.Stdin / os.Stdout)
// keeps the REPL fully testable without a real terminal.
type REPL struct {
	in  io.Reader
	out io.Writer
	env *object.Environment
}

// New creates a REPL ready to use. The environment is created once here and
// shared across every iteration of the loop — this is what makes variables
// defined on one line visible on the next.
func New(in io.Reader, out io.Writer) *REPL {
	return &REPL{
		in:  in,
		out: out,
		env: object.NewEnvironment(),
	}
}

// Start runs the Read-Eval-Print loop until EOF or the user types exit/quit.
// Each iteration:
//  1. Prints the prompt.
//  2. Reads one line of input.
//  3. Runs it through the full Lexer → Parser → Evaluator pipeline.
//  4. Prints the result (if any).
func (r *REPL) Start() {
	scanner := bufio.NewScanner(r.in)
	for {
		fmt.Fprint(r.out, PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			// EOF or read error — exit the loop cleanly.
			return
		}

		line := scanner.Text()

		if line == "exit" || line == "quit" {
			fmt.Fprintln(r.out, "Goodbye!")
			return
		}

		r.eval(line)
	}
}

// eval runs a single line of Zen source through the full pipeline and prints
// the result. Parse errors are reported and the loop continues — the REPL
// never exits due to a user mistake.
func (r *REPL) eval(input string) {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		r.printParseErrors(p.Errors())
		return
	}

	result := evaluator.Eval(program, r.env)

	// nil is returned by let statements — nothing to print.
	// NULL is the explicit null value — printing it on every variable
	// assignment would be noisy, so we suppress it too.
	if result == nil || result == object.NULL {
		return
	}

	fmt.Fprintln(r.out, result.Inspect())
}

// printParseErrors formats and prints all parse errors collected by the parser.
func (r *REPL) printParseErrors(errors []*parser.ParseError) {
	fmt.Fprintln(r.out, "parse errors:")
	for _, err := range errors {
		fmt.Fprintln(r.out, "\t"+err.Error())
	}
}
