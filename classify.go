package plainfields

import (
	"unicode"
)

// Helper functions for character classification
func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch rune) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch rune) bool {
	return ch >= '0' && ch <= '7'
}

func isBinaryDigit(ch rune) bool {
	return ch == '0' || ch == '1'
}

func isStringStart(ch rune) bool {
	return ch == '"' || ch == '\''
}

func isNumericSign(ch rune) bool {
	return ch == '+' || ch == '-'
}

func isIdentifierContinue(ch rune) bool {
	return isLetter(ch) || isDigit(ch) || ch == '-' || ch == '_'
}
