package lexer

// InputReader is a cursor over the source string.
// It tracks two positions:
//   - position     — the character currently being examined
//   - readPosition — always one ahead, used for peeking without consuming
type InputReader struct {
	source       string
	position     int
	readPosition int
}

// NewInputReader creates a ready-to-use InputReader.
// Calls Advance() once to prime it — position lands on index 0, readPosition on index 1.
func NewInputReader(source string) *InputReader {
	ir := &InputReader{source: source}
	ir.Advance()
	return ir
}

// CurrentChar returns the character at the current position.
// Returns null byte (0) when past the end of source.
func (ir *InputReader) CurrentChar() byte {
	if ir.position >= len(ir.source) {
		return 0
	}
	return ir.source[ir.position]
}

// PeekChar returns the character at readPosition without advancing.
// Returns null byte (0) if there is no next character.
func (ir *InputReader) PeekChar() byte {
	if ir.readPosition >= len(ir.source) {
		return 0
	}
	return ir.source[ir.readPosition]
}

// Advance copies readPosition into position, increments readPosition,
// and returns the character now at position.
func (ir *InputReader) Advance() byte {
	ir.position = ir.readPosition
	ir.readPosition++
	return ir.CurrentChar()
}

// IsAtEnd returns true when the cursor has moved past the last valid character.
func (ir *InputReader) IsAtEnd() bool {
	return ir.position >= len(ir.source)
}
