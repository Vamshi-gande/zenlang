package token

import "testing"

func TestLookupIdentifier_Keywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"let", LET},
		{"fn", FUNCTION},
		{"if", IF},
		{"else", ELSE},
		{"return", RETURN},
		{"while", WHILE},
		{"true", TRUE},
		{"false", FALSE},
	}

	for _, tt := range tests {
		got := LookupIdentifier(tt.input)
		if got != tt.expected {
			t.Errorf("LookupIdentifier(%q): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}

func TestLookupIdentifier_Identifiers(t *testing.T) {
	identifiers := []string{
		"x",
		"myVar",
		"letter",   // contains "let" but is NOT the keyword
		"ifelse",   // contains "if"  but is NOT the keyword
		"returned", // contains "return" but is NOT the keyword
		"truthy",   // contains "true" but is NOT the keyword
		"foobar",
		"counter_1",
	}

	for _, ident := range identifiers {
		got := LookupIdentifier(ident)
		if got != IDENT {
			t.Errorf("LookupIdentifier(%q): expected IDENT, got %q", ident, got)
		}
	}
}

func TestLookupIdentifier_CaseSensitive(t *testing.T) {
	nonKeywords := []string{"Let", "LET", "Fn", "FN", "IF", "True", "FALSE", "Return"}

	for _, ident := range nonKeywords {
		got := LookupIdentifier(ident)
		if got != IDENT {
			t.Errorf("LookupIdentifier(%q): expected IDENT (case sensitive), got %q", ident, got)
		}
	}
}
