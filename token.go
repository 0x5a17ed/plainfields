package plainfields

import (
	"fmt"
)

// TokenType identifies the type of lex tokens.
type TokenType int

const (
	TokenError TokenType = iota // Error occurred; value is the text of error.
	TokenEOF                    // End of the file reached.

	TokenFieldPrefix    // `^` or `!`
	TokenIdentifier     // `abc-123`
	TokenAssign         // `=`
	TokenNumber         // `123`, `0xFF`, `0b11`, etc.
	TokenString         // `"abc"` or `'def'`
	TokenTrue           // `true`
	TokenFalse          // `false`
	TokenNil            // `nil`
	TokenFieldSeparator // `,`
	TokenListSeparator  // `;`
	TokenPairSeparator  // `:`
)

func (t TokenType) String() string {
	switch t {
	case TokenError:
		return "Error"
	case TokenEOF:
		return "EOF"
	case TokenFieldPrefix:
		return "FieldPrefix"
	case TokenIdentifier:
		return "Identifier"
	case TokenAssign:
		return "Assign"
	case TokenNumber:
		return "Number"
	case TokenString:
		return "String"
	case TokenTrue:
		return "True"
	case TokenFalse:
		return "False"
	case TokenNil:
		return "Nil"
	case TokenFieldSeparator:
		return "FieldSeparator"
	case TokenListSeparator:
		return "ListSeparator"
	case TokenPairSeparator:
		return "PairSeparator"
	default:
		return fmt.Sprintf("TokenType(%d)", t)
	}
}

// Token represents a Token produced by the lexer.
type Token struct {
	Typ TokenType // Type of this Token.
	Pos Position  // Starting Position of the Token in the input.
	Val string    // Token text.
}

func (t Token) String() string {
	return fmt.Sprintf("{%s at %s: %+#q}", t.Typ, t.Pos, t.Val)
}
