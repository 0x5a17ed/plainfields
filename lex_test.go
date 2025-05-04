package plainfields

import (
	"reflect"
	"slices"
	"testing"
)

func TestLex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{"empty input", "", []Token{
			{Typ: TokenEOF, Pos: Position{Offset: 0, Column: 1}, Val: ""},
		}},
		{"whitespace only", "   \t\n  ", []Token{
			{Typ: TokenEOF, Pos: Position{Offset: 7, Column: 8}, Val: ""},
		}},
		{"simple identifier", "abc=1", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "abc"},
			{Typ: TokenAssign, Pos: Position{Offset: 3, Column: 4}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 4, Column: 5}, Val: "1"},
			{Typ: TokenEOF, Pos: Position{Offset: 5, Column: 6}, Val: ""},
		}},
		{"identifier with hyphen", "abc-123=1", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "abc-123"},
			{Typ: TokenAssign, Pos: Position{Offset: 7, Column: 8}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 8, Column: 9}, Val: "1"},
			{Typ: TokenEOF, Pos: Position{Offset: 9, Column: 10}, Val: ""},
		}},
		{"identifier with underscore", "abc_123=1", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "abc_123"},
			{Typ: TokenAssign, Pos: Position{Offset: 7, Column: 8}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 8, Column: 9}, Val: "1"},
			{Typ: TokenEOF, Pos: Position{Offset: 9, Column: 10}, Val: ""},
		}},
		{"field prefix ^", "^abc", []Token{
			{Typ: TokenBooleanPrefix, Pos: Position{Offset: 0, Column: 1}, Val: "^"},
			{Typ: TokenIdentifier, Pos: Position{Offset: 1, Column: 2}, Val: "abc"},
			{Typ: TokenEOF, Pos: Position{Offset: 4, Column: 5}, Val: ""},
		}},
		{"field prefix !", "!xyz", []Token{
			{Typ: TokenBooleanPrefix, Pos: Position{Offset: 0, Column: 1}, Val: "!"},
			{Typ: TokenIdentifier, Pos: Position{Offset: 1, Column: 2}, Val: "xyz"},
			{Typ: TokenEOF, Pos: Position{Offset: 4, Column: 5}, Val: ""},
		}},
		{"value binding with single value", "name=John", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "name"},
			{Typ: TokenAssign, Pos: Position{Offset: 4, Column: 5}, Val: "="},
			{Typ: TokenIdentifier, Pos: Position{Offset: 5, Column: 6}, Val: "John"},
			{Typ: TokenEOF, Pos: Position{Offset: 9, Column: 10}, Val: ""},
		}},
		{"value binding with multiple values", "colors=red;blue;green", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "colors"},
			{Typ: TokenAssign, Pos: Position{Offset: 6, Column: 7}, Val: "="},
			{Typ: TokenIdentifier, Pos: Position{Offset: 7, Column: 8}, Val: "red"},
			{Typ: TokenListSeparator, Pos: Position{Offset: 10, Column: 11}, Val: ";"},
			{Typ: TokenIdentifier, Pos: Position{Offset: 11, Column: 12}, Val: "blue"},
			{Typ: TokenListSeparator, Pos: Position{Offset: 15, Column: 16}, Val: ";"},
			{Typ: TokenIdentifier, Pos: Position{Offset: 16, Column: 17}, Val: "green"},
			{Typ: TokenEOF, Pos: Position{Offset: 21, Column: 22}, Val: ""},
		}},
		{"value binding with a map value", `values=a:1;"b":2;'c':true;0:false`, []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "values"},
			{Typ: TokenAssign, Pos: Position{Offset: 6, Column: 7}, Val: "="},
			{Typ: TokenIdentifier, Pos: Position{Offset: 7, Column: 8}, Val: "a"},
			{Typ: TokenPairSeparator, Pos: Position{Offset: 8, Column: 9}, Val: ":"},
			{Typ: TokenNumber, Pos: Position{Offset: 9, Column: 10}, Val: "1"},
			{Typ: TokenListSeparator, Pos: Position{Offset: 10, Column: 11}, Val: ";"},
			{Typ: TokenString, Pos: Position{Offset: 11, Column: 12}, Val: `"b"`},
			{Typ: TokenPairSeparator, Pos: Position{Offset: 14, Column: 15}, Val: ":"},
			{Typ: TokenNumber, Pos: Position{Offset: 15, Column: 16}, Val: "2"},
			{Typ: TokenListSeparator, Pos: Position{Offset: 16, Column: 17}, Val: ";"},
			{Typ: TokenString, Pos: Position{Offset: 17, Column: 18}, Val: `'c'`},
			{Typ: TokenPairSeparator, Pos: Position{Offset: 20, Column: 21}, Val: ":"},
			{Typ: TokenTrue, Pos: Position{Offset: 21, Column: 22}, Val: "true"},
			{Typ: TokenListSeparator, Pos: Position{Offset: 25, Column: 26}, Val: ";"},
			{Typ: TokenNumber, Pos: Position{Offset: 26, Column: 27}, Val: "0"},
			{Typ: TokenPairSeparator, Pos: Position{Offset: 27, Column: 28}, Val: ":"},
			{Typ: TokenFalse, Pos: Position{Offset: 28, Column: 29}, Val: "false"},
			{Typ: TokenEOF, Pos: Position{Offset: 33, Column: 34}, Val: ""},
		}},
		{"multiple values with multiple values", "colors=red;blue;green,shapes=circle;square;triangle", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "colors"},
			{Typ: TokenAssign, Pos: Position{Offset: 6, Column: 7}, Val: "="},
			{Typ: TokenIdentifier, Pos: Position{Offset: 7, Column: 8}, Val: "red"},
			{Typ: TokenListSeparator, Pos: Position{Offset: 10, Column: 11}, Val: ";"},
			{Typ: TokenIdentifier, Pos: Position{Offset: 11, Column: 12}, Val: "blue"},
			{Typ: TokenListSeparator, Pos: Position{Offset: 15, Column: 16}, Val: ";"},
			{Typ: TokenIdentifier, Pos: Position{Offset: 16, Column: 17}, Val: "green"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 21, Column: 22}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 22, Column: 23}, Val: "shapes"},
			{Typ: TokenAssign, Pos: Position{Offset: 28, Column: 29}, Val: "="},
			{Typ: TokenIdentifier, Pos: Position{Offset: 29, Column: 30}, Val: "circle"},
			{Typ: TokenListSeparator, Pos: Position{Offset: 35, Column: 36}, Val: ";"},
			{Typ: TokenIdentifier, Pos: Position{Offset: 36, Column: 37}, Val: "square"},
			{Typ: TokenListSeparator, Pos: Position{Offset: 42, Column: 43}, Val: ";"},
			{Typ: TokenIdentifier, Pos: Position{Offset: 43, Column: 44}, Val: "triangle"},
			{Typ: TokenEOF, Pos: Position{Offset: 51, Column: 52}, Val: ""},
		}},
		{"multiple fields", "name=John,age=30,active=true", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "name"},
			{Typ: TokenAssign, Pos: Position{Offset: 4, Column: 5}, Val: "="},
			{Typ: TokenIdentifier, Pos: Position{Offset: 5, Column: 6}, Val: "John"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 9, Column: 10}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 10, Column: 11}, Val: "age"},
			{Typ: TokenAssign, Pos: Position{Offset: 13, Column: 14}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 14, Column: 15}, Val: "30"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 16, Column: 17}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 17, Column: 18}, Val: "active"},
			{Typ: TokenAssign, Pos: Position{Offset: 23, Column: 24}, Val: "="},
			{Typ: TokenTrue, Pos: Position{Offset: 24, Column: 25}, Val: "true"},
			{Typ: TokenEOF, Pos: Position{Offset: 28, Column: 29}, Val: ""},
		}},
		{"keywords", "enabled=true,disabled=false,ptr=nil", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "enabled"},
			{Typ: TokenAssign, Pos: Position{Offset: 7, Column: 8}, Val: "="},
			{Typ: TokenTrue, Pos: Position{Offset: 8, Column: 9}, Val: "true"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 12, Column: 13}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 13, Column: 14}, Val: "disabled"},
			{Typ: TokenAssign, Pos: Position{Offset: 21, Column: 22}, Val: "="},
			{Typ: TokenFalse, Pos: Position{Offset: 22, Column: 23}, Val: "false"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 27, Column: 28}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 28, Column: 29}, Val: "ptr"},
			{Typ: TokenAssign, Pos: Position{Offset: 31, Column: 32}, Val: "="},
			{Typ: TokenNil, Pos: Position{Offset: 32, Column: 33}, Val: "nil"},
			{Typ: TokenEOF, Pos: Position{Offset: 35, Column: 36}, Val: ""},
		}},
		{"decimal numbers", "a=123,b=-456,c=3.14,d=1e5,e=2.5e-3,f=0.123,g=0_1", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "a"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 2, Column: 3}, Val: "123"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 5, Column: 6}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 6, Column: 7}, Val: "b"},
			{Typ: TokenAssign, Pos: Position{Offset: 7, Column: 8}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 8, Column: 9}, Val: "-456"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 12, Column: 13}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 13, Column: 14}, Val: "c"},
			{Typ: TokenAssign, Pos: Position{Offset: 14, Column: 15}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 15, Column: 16}, Val: "3.14"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 19, Column: 20}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 20, Column: 21}, Val: "d"},
			{Typ: TokenAssign, Pos: Position{Offset: 21, Column: 22}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 22, Column: 23}, Val: "1e5"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 25, Column: 26}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 26, Column: 27}, Val: "e"},
			{Typ: TokenAssign, Pos: Position{Offset: 27, Column: 28}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 28, Column: 29}, Val: "2.5e-3"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 34, Column: 35}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 35, Column: 36}, Val: "f"},
			{Typ: TokenAssign, Pos: Position{Offset: 36, Column: 37}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 37, Column: 38}, Val: "0.123"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 42, Column: 43}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 43, Column: 44}, Val: "g"},
			{Typ: TokenAssign, Pos: Position{Offset: 44, Column: 45}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 45, Column: 46}, Val: "0_1"},
			{Typ: TokenEOF, Pos: Position{Offset: 48, Column: 49}, Val: ""},
		}},
		{"hex numbers", "h1=0xFF,h2=0x123.4p5,h3=0x1p1,h4=0x123.4", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "h1"},
			{Typ: TokenAssign, Pos: Position{Offset: 2, Column: 3}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 3, Column: 4}, Val: "0xFF"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 7, Column: 8}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 8, Column: 9}, Val: "h2"},
			{Typ: TokenAssign, Pos: Position{Offset: 10, Column: 11}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 11, Column: 12}, Val: "0x123.4p5"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 20, Column: 21}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 21, Column: 22}, Val: "h3"},
			{Typ: TokenAssign, Pos: Position{Offset: 23, Column: 24}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 24, Column: 25}, Val: "0x1p1"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 29, Column: 30}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 30, Column: 31}, Val: "h4"},
			{Typ: TokenAssign, Pos: Position{Offset: 32, Column: 33}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 33, Column: 34}, Val: "0x123.4"},
			{Typ: TokenEOF, Pos: Position{Offset: 40, Column: 41}, Val: ""},
		}},
		{"octal and binary numbers", "o=0o777,b=0b1101", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "o"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 2, Column: 3}, Val: "0o777"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 7, Column: 8}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 8, Column: 9}, Val: "b"},
			{Typ: TokenAssign, Pos: Position{Offset: 9, Column: 10}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 10, Column: 11}, Val: "0b1101"},
			{Typ: TokenEOF, Pos: Position{Offset: 16, Column: 17}, Val: ""},
		}},
		{"string literals", `s1="hello",s2='world'`, []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "s1"},
			{Typ: TokenAssign, Pos: Position{Offset: 2, Column: 3}, Val: "="},
			{Typ: TokenString, Pos: Position{Offset: 3, Column: 4}, Val: `"hello"`},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 10, Column: 11}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 11, Column: 12}, Val: "s2"},
			{Typ: TokenAssign, Pos: Position{Offset: 13, Column: 14}, Val: "="},
			{Typ: TokenString, Pos: Position{Offset: 14, Column: 15}, Val: `'world'`},
			{Typ: TokenEOF, Pos: Position{Offset: 21, Column: 22}, Val: ""},
		}},
		{"string with escapes", `s="a\nb\tc\"d"`, []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "s"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenString, Pos: Position{Offset: 2, Column: 3}, Val: `"a\nb\tc\"d"`},
			{Typ: TokenEOF, Pos: Position{Offset: 14, Column: 15}, Val: ""},
		}},
		{"whitespace handling", "  a  =  123  ,  b  =  true  ", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 2, Column: 3}, Val: "a"},
			{Typ: TokenAssign, Pos: Position{Offset: 5, Column: 6}, Val: "="},
			{Typ: TokenNumber, Pos: Position{Offset: 8, Column: 9}, Val: "123"},
			{Typ: TokenFieldSeparator, Pos: Position{Offset: 13, Column: 14}, Val: ","},
			{Typ: TokenIdentifier, Pos: Position{Offset: 16, Column: 17}, Val: "b"},
			{Typ: TokenAssign, Pos: Position{Offset: 19, Column: 20}, Val: "="},
			{Typ: TokenTrue, Pos: Position{Offset: 22, Column: 23}, Val: "true"},
			{Typ: TokenEOF, Pos: Position{Offset: 28, Column: 29}, Val: ""},
		}},

		// Error cases.
		{"error: unexpected character", "a=@", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "a"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenError, Pos: Position{Offset: 2, Column: 3}, Val: "unexpected character: U+0040 '@'"},
		}},
		{"error: unterminated string", `s="hello`, []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "s"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenError, Pos: Position{Offset: 2, Column: 3}, Val: "unterminated string"},
		}},
		{"error: unterminated escape sequence", `s="hello\`, []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "s"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenError, Pos: Position{Offset: 2, Column: 3}, Val: "unterminated escape sequence"},
		}},
		{"error: exponent missing digits", "e=1e", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "e"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenError, Pos: Position{Offset: 2, Column: 3}, Val: "expected digit after exponent"},
		}},
		{"error: invalid hex number", "h=0xGHI", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "h"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenError, Pos: Position{Offset: 2, Column: 3}, Val: "expected hex digit"},
		}},
		{"error: invalid hex exponent", "h=0x1p+G", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "h"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenError, Pos: Position{Offset: 2, Column: 3}, Val: "expected digit after hex exponent"},
		}},
		{"error: invalid octal number", "o=0o8", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "o"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenError, Pos: Position{Offset: 2, Column: 3}, Val: "expected octal digit"},
		}},
		{"error: invalid binary number", "b=0b2", []Token{
			{Typ: TokenIdentifier, Pos: Position{Offset: 0, Column: 1}, Val: "b"},
			{Typ: TokenAssign, Pos: Position{Offset: 1, Column: 2}, Val: "="},
			{Typ: TokenError, Pos: Position{Offset: 2, Column: 3}, Val: "expected binary digit"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Collect(Lex(tt.input))

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Lex(%q) =\n  got:  %v\n  want: %v", tt.input, got, tt.expected)
			}
		})
	}
}
