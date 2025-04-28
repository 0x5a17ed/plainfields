package plainfields

import (
	"fmt"
	"testing"
)

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{TokenError, "Error"},
		{TokenEOF, "EOF"},
		{TokenFieldPrefix, "FieldPrefix"},
		{TokenIdentifier, "Identifier"},
		{TokenAssign, "Assign"},
		{TokenNumber, "Number"},
		{TokenString, "String"},
		{TokenTrue, "True"},
		{TokenFalse, "False"},
		{TokenNil, "Nil"},
		{TokenFieldSeparator, "FieldSeparator"},
		{TokenListSeparator, "ListSeparator"},
		{TokenPairSeparator, "PairSeparator"},

		// The silly part: test invalid token types
		{TokenType(9999), "TokenType(9999)"},
		{TokenType(-42), "TokenType(-42)"},
		{TokenType(1337), "TokenType(1337)"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("TokenType(%d)", tt.tokenType), func(t *testing.T) {
			result := tt.tokenType.String()
			if result != tt.expected {
				t.Errorf("expected: %q, got: %q", tt.expected, result)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	token := Token{
		Typ: TokenIdentifier,
		Pos: Position{Offset: 42, Column: 7},
		Val: "test-value",
	}

	expected := "{Identifier at Col 7 (Offset 42): `test-value`}"
	result := token.String()

	if result != expected {
		t.Errorf("expected: %q, got: %q", expected, result)
	}

	// Test some silly tokens for fun.
	sillyTokens := []Token{
		{TokenError, Position{Offset: 0, Column: 1}, "oopsie!"},
		{TokenString, Position{Offset: 123, Column: 456}, "\"escaped\\quotes\""},
		{TokenNumber, Position{Offset: 42, Column: 42}, "0xDEADBEEF"},
		{TokenType(9999), Position{Offset: 9999, Column: 9999}, "👽"},
	}

	for i, token := range sillyTokens {
		t.Run(fmt.Sprintf("SillyToken_%d", i), func(t *testing.T) {
			result := token.String()
			if result == "" {
				t.Error("expected non-empty string")
			}
		})
	}
}
