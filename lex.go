package plainfields

import (
	"fmt"
	"iter"
	"unicode/utf8"
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the advance state.
type stateFn func(*lexer) stateFn

// lexer holds the state of our scanner.
type lexer struct {
	input string // The string being scanned.

	yield func(Token) bool // Yield callback.
	done  bool             // Set to true if yield returns false.

	start Position // Start Position of the current Token.
	pos   Position // Current Position in the input.
	prev  Position // Previous Position (for undo).
}

// next returns the next rune in the input and updates the lexer's Position.
// It saves the current Position to allow undo.
func (l *lexer) next() rune {
	// Save the current state for undo.
	l.prev = l.pos

	// Check if we are at the end of the input.
	if l.pos.Offset >= len(l.input) {
		return eof
	}

	// Advance to the advance rune.
	r, w := utf8.DecodeRuneInString(l.input[l.pos.Offset:])
	l.pos.Offset += w

	// Update column.
	l.pos.Column++

	return r
}

// undo reverts the lexer by one rune using the saved Position.
func (l *lexer) undo() {
	l.pos = l.prev
}

// peek returns but does not consume the next rune.
func (l *lexer) peek() rune {
	r := l.next()
	l.undo()
	return r
}

// text returns the text of the current Token.
func (l *lexer) text() string {
	if l.start.Offset == l.pos.Offset {
		return ""
	}
	return l.input[l.start.Offset:l.pos.Offset]
}

// emit creates a Token from the current input and calls the yield callback.
func (l *lexer) emit(typ TokenType) {
	if l.done || !l.yield(Token{Typ: typ, Pos: l.start, Val: l.text()}) {
		l.done = true
	}

	// Set the start of the advance Token.
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// errorf emits an error Token and stops lexing.
func (l *lexer) errorf(format string, args ...any) stateFn {
	msg := fmt.Sprintf(format, args...)
	l.yield(Token{Typ: TokenError, Pos: l.start, Val: msg})
	l.done = true
	return nil
}

func lexTop(l *lexer) stateFn {
	switch ch := l.peek(); {
	case ch == eof:
		l.ignore()
		l.emit(TokenEOF)
		return nil
	case isSpace(ch):
		l.next()
		l.ignore()
		return lexTop

	case ch == '^' || ch == '!':
		l.next()
		l.emit(TokenBooleanPrefix)
		return lexTop
	case ch == '=':
		l.next()
		l.emit(TokenAssign)
		return lexTop
	case ch == ',':
		l.next()
		l.emit(TokenFieldSeparator)
		return lexTop
	case ch == ';':
		l.next()
		l.emit(TokenListSeparator)
		return lexTop
	case ch == ':':
		l.next()
		l.emit(TokenPairSeparator)
		return lexTop

	case isStringStart(ch):
		return lexString
	case isNumericSign(ch) || isDigit(ch):
		return lexNumber
	case isLetter(ch):
		l.next()
		return lexIdentifierOrKeywordContinue

	default:
		return l.errorf("unexpected character: %#U", ch)
	}
}

func lexIdentifierOrKeywordContinue(l *lexer) stateFn {
	if isIdentifierContinue(l.peek()) {
		l.next()
		return lexIdentifierOrKeywordContinue
	}

	switch text := l.text(); text {
	case "true":
		l.emit(TokenTrue)
		return lexTop
	case "false":
		l.emit(TokenFalse)
		return lexTop
	case "nil":
		l.emit(TokenNil)
		return lexTop
	default:
		l.emit(TokenIdentifier)
		return lexTop
	}
}

func lexString(l *lexer) stateFn {
	// Get the opening quote
	quote := l.next()

	var fn stateFn
	fn = func(l *lexer) stateFn {
		return lexStringContent(l, quote, fn)
	}

	return fn
}

func lexStringContent(l *lexer, quote rune, fn stateFn) stateFn {
	switch ch := l.next(); {
	case ch == eof:
		return l.errorf("unterminated string")
	case ch == quote:
		l.emit(TokenString)
		return lexTop
	case ch == '\\':
		return lexStringEscape(l, fn)
	default:
		return fn
	}
}

func lexStringEscape(l *lexer, fn stateFn) stateFn {
	if l.next() == eof {
		return l.errorf("unterminated escape sequence")
	}
	return fn
}

func lexNumber(l *lexer) stateFn {
	// Handle optional sign.
	if isNumericSign(l.peek()) {
		l.next()
	}

	// Check if it's a special number (hex, octal, binary)
	if l.peek() == '0' {
		l.next()
		switch ch := l.peek(); {
		case ch == 'x' || ch == 'X':
			l.next()
			return lexHexDigits
		case ch == 'o' || ch == 'O':
			l.next()
			return lexOctalDigits
		case ch == 'b' || ch == 'B':
			l.next()
			return lexBinaryDigits
		case isDigit(ch) || ch == '_':
			return lexDecimalDigits
		case ch == '.':
			return lexDecimalFraction
		}
		l.emit(TokenNumber) // Just "0"
		return lexTop
	}

	return lexDecimalDigits
}

func lexDecimalDigits(l *lexer) stateFn {
	ch := l.peek()
	if isDigit(ch) || ch == '_' {
		l.next()
		return lexDecimalDigits
	}
	if ch == '.' {
		return lexDecimalFraction
	}
	if ch == 'e' || ch == 'E' {
		return lexExponent
	}
	l.emit(TokenNumber)
	return lexTop
}

func lexDecimalFraction(l *lexer) stateFn {
	l.next() // consume '.'
	return lexDecimalFractionDigits
}

func lexDecimalFractionDigits(l *lexer) stateFn {
	ch := l.peek()
	if isDigit(ch) || ch == '_' {
		l.next()
		return lexDecimalFractionDigits
	}
	if ch == 'e' || ch == 'E' {
		return lexExponent
	}
	l.emit(TokenNumber)
	return lexTop
}

func lexExponent(l *lexer) stateFn {
	l.next() // consume 'e' or 'E'

	// Handle optional sign.
	if isNumericSign(l.peek()) {
		l.next()
	}

	// Must have at least one digit
	if !isDigit(l.peek()) {
		return l.errorf("expected digit after exponent")
	}

	return lexExponentDigits
}

func lexExponentDigits(l *lexer) stateFn {
	ch := l.peek()
	if isDigit(ch) || ch == '_' {
		l.next()
		return lexExponentDigits
	}
	l.emit(TokenNumber)
	return lexTop
}

func lexHexDigits(l *lexer) stateFn {
	// Must have at least one hex digit
	if !isHexDigit(l.peek()) {
		return l.errorf("expected hex digit")
	}
	return lexHexDigitsContinue
}

func lexHexDigitsContinue(l *lexer) stateFn {
	ch := l.peek()
	if isHexDigit(ch) || ch == '_' {
		l.next()
		return lexHexDigitsContinue
	}
	if ch == '.' {
		return lexHexFraction
	}
	if ch == 'p' || ch == 'P' {
		return lexHexExponent
	}
	l.emit(TokenNumber)
	return lexTop
}

func lexHexFraction(l *lexer) stateFn {
	l.next() // consume '.'
	return lexHexFractionDigits
}

func lexHexFractionDigits(l *lexer) stateFn {
	ch := l.peek()
	if isHexDigit(ch) || ch == '_' {
		l.next()
		return lexHexFractionDigits
	}
	if ch == 'p' || ch == 'P' {
		return lexHexExponent
	}
	l.emit(TokenNumber)
	return lexTop
}

func lexHexExponent(l *lexer) stateFn {
	l.next() // consume 'p' or 'P'

	// Handle optional sign.
	if isNumericSign(l.peek()) {
		l.next()
	}

	// Must have at least one digit
	if !isDigit(l.peek()) {
		return l.errorf("expected digit after hex exponent")
	}

	return lexHexExponentDigits
}

func lexHexExponentDigits(l *lexer) stateFn {
	ch := l.peek()
	if isDigit(ch) || ch == '_' {
		l.next()
		return lexHexExponentDigits
	}
	l.emit(TokenNumber)
	return lexTop
}

func lexOctalDigits(l *lexer) stateFn {
	// Must have at least one octal digit
	if !isOctalDigit(l.peek()) {
		return l.errorf("expected octal digit")
	}
	return lexOctalDigitsContinue
}

func lexOctalDigitsContinue(l *lexer) stateFn {
	ch := l.peek()
	if isOctalDigit(ch) || ch == '_' {
		l.next()
		return lexOctalDigitsContinue
	}
	l.emit(TokenNumber)
	return lexTop
}

func lexBinaryDigits(l *lexer) stateFn {
	// Must have at least one binary digit
	if !isBinaryDigit(l.peek()) {
		return l.errorf("expected binary digit")
	}
	return lexBinaryDigitsContinue
}

func lexBinaryDigitsContinue(l *lexer) stateFn {
	ch := l.peek()
	if isBinaryDigit(ch) || ch == '_' {
		l.next()
		return lexBinaryDigitsContinue
	}
	l.emit(TokenNumber)
	return lexTop
}

func runPattern(l *lexer) {
	for state := lexTop; state != nil; state = state(l) {
		if l.done {
			break
		}
	}
}

// Lex returns a lazy iterator lexer for the input yielding tokens.
func Lex(input string) iter.Seq[Token] {
	return func(yield func(Token) bool) {
		l := &lexer{
			input: input,
			yield: yield,

			// Initialize positions: starting at offset 0, line 1, column 1.
			start: Position{Offset: 0, Column: 1},
			pos:   Position{Offset: 0, Column: 1},
			prev:  Position{Offset: 0, Column: 1},
		}
		runPattern(l)
	}
}
