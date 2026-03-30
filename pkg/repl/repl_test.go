package repl

import (
	"bytes"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helper
// ---------------------------------------------------------------------------

// runREPL feeds the given input string to a fresh REPL, runs Start(), and
// returns the captured output. Input lines should be newline-separated.
func runREPL(input string) string {
	in := strings.NewReader(input)
	out := &bytes.Buffer{}
	r := New(in, out)
	r.Start()
	return out.String()
}

// ---------------------------------------------------------------------------
// Basic expression evaluation
// ---------------------------------------------------------------------------

func TestSingleExpression(t *testing.T) {
	output := runREPL("5 + 3\n")
	if !strings.Contains(output, "8") {
		t.Errorf("expected output to contain '8', got: %q", output)
	}
}

func TestArithmeticPrecedence(t *testing.T) {
	output := runREPL("2 + 3 * 4\n")
	if !strings.Contains(output, "14") {
		t.Errorf("expected output to contain '14', got: %q", output)
	}
}

func TestBooleanExpression(t *testing.T) {
	output := runREPL("5 > 3\n")
	if !strings.Contains(output, "true") {
		t.Errorf("expected output to contain 'true', got: %q", output)
	}
}

func TestStringExpression(t *testing.T) {
	output := runREPL(`"hello" + " " + "world"` + "\n")
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected output to contain 'hello world', got: %q", output)
	}
}

// ---------------------------------------------------------------------------
// Variable persistence across lines
// ---------------------------------------------------------------------------

func TestVariablePersistenceAcrossLines(t *testing.T) {
	// x is defined on the first line and used on the second.
	// The shared environment must carry it across iterations.
	output := runREPL("let x = 10\nx + 5\n")
	if !strings.Contains(output, "15") {
		t.Errorf("expected output to contain '15', got: %q", output)
	}
}

func TestMultipleVariablesPersist(t *testing.T) {
	output := runREPL("let a = 3\nlet b = 7\na * b\n")
	if !strings.Contains(output, "21") {
		t.Errorf("expected output to contain '21', got: %q", output)
	}
}

func TestFunctionDefinitionAndCall(t *testing.T) {
	input := "let add = fn(a, b) { a + b }\nadd(4, 6)\n"
	output := runREPL(input)
	if !strings.Contains(output, "10") {
		t.Errorf("expected output to contain '10', got: %q", output)
	}
}

func TestClosurePersists(t *testing.T) {
	input := "let makeAdder = fn(x) { fn(y) { x + y } }\nlet addFive = makeAdder(5)\naddFive(3)\n"
	output := runREPL(input)
	if !strings.Contains(output, "8") {
		t.Errorf("expected output to contain '8', got: %q", output)
	}
}

// ---------------------------------------------------------------------------
// NULL suppression
// ---------------------------------------------------------------------------

// let statements return nil/NULL — the REPL should produce no visible output
// for them (just the prompt). Printing "null" on every variable binding would
// be confusing and noisy.
func TestNullNotPrintedForLetStatement(t *testing.T) {
	output := runREPL("let x = 5\n")
	if strings.Contains(output, "null") || strings.Contains(output, "NULL") {
		t.Errorf("expected 'null' NOT to appear in output for let statement, got: %q", output)
	}
}

func TestNullNotPrintedForFunctionDefinition(t *testing.T) {
	output := runREPL("let f = fn(x) { x }\n")
	if strings.Contains(output, "null") || strings.Contains(output, "NULL") {
		t.Errorf("expected 'null' NOT to appear for function definition, got: %q", output)
	}
}

// ---------------------------------------------------------------------------
// Parse errors
// ---------------------------------------------------------------------------

// Parse errors must be reported clearly and the REPL must continue running —
// it should not crash or exit on user mistakes.
func TestParseErrorIsReported(t *testing.T) {
	output := runREPL("let = 5\n")
	if !strings.Contains(output, "parse errors:") {
		t.Errorf("expected 'parse errors:' in output, got: %q", output)
	}
}

func TestREPLContinuesAfterParseError(t *testing.T) {
	// First line is invalid, second is valid — the result of the second
	// line must still appear in the output.
	output := runREPL("let = 5\n1 + 1\n")
	if !strings.Contains(output, "parse errors:") {
		t.Errorf("expected parse error to be reported, got: %q", output)
	}
	if !strings.Contains(output, "2") {
		t.Errorf("expected REPL to continue and evaluate '1+1' → 2, got: %q", output)
	}
}

// ---------------------------------------------------------------------------
// Runtime errors
// ---------------------------------------------------------------------------

func TestRuntimeErrorIsDisplayed(t *testing.T) {
	output := runREPL("5 + true\n")
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected 'ERROR' in output for type mismatch, got: %q", output)
	}
}

func TestREPLContinuesAfterRuntimeError(t *testing.T) {
	// Error on line 1, valid expression on line 2.
	output := runREPL("5 + true\n3 * 3\n")
	if !strings.Contains(output, "9") {
		t.Errorf("expected REPL to continue after runtime error and evaluate '3*3' → 9, got: %q", output)
	}
}

func TestUndefinedVariableError(t *testing.T) {
	output := runREPL("foobar\n")
	if !strings.Contains(output, "identifier not found") {
		t.Errorf("expected 'identifier not found' error, got: %q", output)
	}
}

// ---------------------------------------------------------------------------
// Exit / quit commands
// ---------------------------------------------------------------------------

func TestExitCommand(t *testing.T) {
	output := runREPL("exit\n")
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("expected 'Goodbye!' on exit, got: %q", output)
	}
}

func TestQuitCommand(t *testing.T) {
	output := runREPL("quit\n")
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("expected 'Goodbye!' on quit, got: %q", output)
	}
}

func TestExitStopsEvaluation(t *testing.T) {
	// Lines after exit should not be evaluated.
	output := runREPL("exit\n99 + 1\n")
	if strings.Contains(output, "100") {
		t.Errorf("expected REPL to stop at exit, but '100' appeared in output: %q", output)
	}
}

// ---------------------------------------------------------------------------
// Prompt
// ---------------------------------------------------------------------------

func TestPromptIsPresent(t *testing.T) {
	output := runREPL("1\n")
	if !strings.Contains(output, PROMPT) {
		t.Errorf("expected prompt '%s' in output, got: %q", PROMPT, output)
	}
}

func TestPromptAppearsOnEveryLine(t *testing.T) {
	output := runREPL("1\n2\n3\n")
	count := strings.Count(output, PROMPT)
	if count < 3 {
		t.Errorf("expected at least 3 prompts for 3 lines of input, got %d in: %q", count, output)
	}
}

// ---------------------------------------------------------------------------
// Data structures
// ---------------------------------------------------------------------------

func TestArrayLiteral(t *testing.T) {
	output := runREPL("[1, 2, 3]\n")
	if !strings.Contains(output, "[1, 2, 3]") {
		t.Errorf("expected '[1, 2, 3]' in output, got: %q", output)
	}
}

func TestHashLiteral(t *testing.T) {
	output := runREPL(`{"x": 1}` + "\n")
	if !strings.Contains(output, "x") || !strings.Contains(output, "1") {
		t.Errorf("expected hash contents in output, got: %q", output)
	}
}

func TestBuiltinLen(t *testing.T) {
	output := runREPL(`len("hello")` + "\n")
	if !strings.Contains(output, "5") {
		t.Errorf("expected '5' from len(\"hello\"), got: %q", output)
	}
}

// ---------------------------------------------------------------------------
// EOF handling
// ---------------------------------------------------------------------------

func TestEOFExitsGracefully(t *testing.T) {
	// Empty input — scanner immediately returns EOF.
	// Start() must return without panicking.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("REPL panicked on EOF: %v", r)
		}
	}()
	runREPL("")
}
