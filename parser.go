package plainfields

import (
	"fmt"
	"iter"
)

// ParserEvent represents different events that can occur during parsing.
type ParserEvent interface {
	isParserEvent()
}

type (
	// FieldStartEvent represents the beginning of a field.
	FieldStartEvent struct {
		Name string
	}
	// FieldEndEvent represents the end of a field.
	FieldEndEvent struct{}

	// ValueEvent represents a value in the field.
	ValueEvent struct {
		Type  TokenType
		Value string
	}

	// ListStartEvent represents the beginning of a value list.
	ListStartEvent struct{}
	// ListEndEvent represents the end of a value list.
	ListEndEvent struct{}

	// MapStartEvent represents the beginning of a map.
	MapStartEvent struct{}
	// MapEndEvent represents the end of a map.
	MapEndEvent struct{}
	// MapKeyEvent represents a key in a map.
	MapKeyEvent struct {
		Type  TokenType
		Value string
	}

	// ErrorEvent represents an error during parsing.
	ErrorEvent struct {
		Pos Position
		Msg string
	}
)

func (FieldStartEvent) isParserEvent() {}
func (FieldEndEvent) isParserEvent()   {}
func (ValueEvent) isParserEvent()      {}
func (ListStartEvent) isParserEvent()  {}
func (ListEndEvent) isParserEvent()    {}
func (MapStartEvent) isParserEvent()   {}
func (MapEndEvent) isParserEvent()     {}
func (MapKeyEvent) isParserEvent()     {}
func (ErrorEvent) isParserEvent()      {}

// Parser holds the state for parsing.
type Parser struct {
	current  Token
	hasToken bool
	yield    func(ParserEvent) bool
	done     bool
	next     func() (Token, bool)
}

// emit sends an event through yield.
func (p *Parser) emit(event ParserEvent) {
	if p.done || !p.yield(event) {
		p.done = true
	}
}

// errorf sends an error event through yield.
func (p *Parser) errorf(format string, args ...any) bool {
	p.emit(ErrorEvent{
		Pos: p.current.Pos,
		Msg: fmt.Sprintf(format, args...),
	})
	return false
}

// advance advances to the next token.
func (p *Parser) advance() bool {
	if tok, ok := p.next(); ok {
		p.hasToken = true
		p.current = tok
	} else {
		p.hasToken = false
	}

	return p.hasToken
}

// expect checks if the current token is of the expected type
func (p *Parser) expect(typ TokenType) bool {
	if !p.hasToken {
		return p.errorf("unexpected end of input")
	}
	if p.current.Typ != typ {
		return p.errorf("expected %s, got %s", typ, p.current.Typ)
	}
	return true
}

// parseFieldList parses the top-level field list
func (p *Parser) parseFieldList() {
	for !p.done && p.advance() {
		if p.current.Typ == TokenEOF {
			break
		}

		if !p.parseField() {
			return
		}

		// If there's a field separator, consume it.
		if p.hasToken && p.current.Typ == TokenFieldSeparator {
			continue
		}
	}
}

// parseField parses a single field
func (p *Parser) parseField() bool {
	switch p.current.Typ {
	case TokenFieldPrefix:
		prefix := p.current.Val // '^' or '!'
		if !p.advance() || !p.expect(TokenIdentifier) {
			return false
		}
		name := p.current.Val

		// Emit as a boolean assignment.
		p.emit(FieldStartEvent{name})
		p.emit(ListStartEvent{})
		if prefix == "^" {
			p.emit(ValueEvent{TokenTrue, "true"})
		} else { // `!`
			p.emit(ValueEvent{TokenFalse, "false"})
		}
		p.emit(ListEndEvent{})
		p.emit(FieldEndEvent{})
		p.advance()

	case TokenIdentifier:
		p.emit(FieldStartEvent{Name: p.current.Val})

		if !p.advance() || !p.expect(TokenAssign) {
			return false
		}

		if !p.advance() || !p.parseValueList() {
			return false
		}

		p.emit(FieldEndEvent{})

	default:
		return p.errorf("expected field prefix or identifier, got %s", p.current.Typ)
	}

	return true
}

// parseValueList parses a list of values, detecting if it's a map.
func (p *Parser) parseValueList() bool {
	if !p.parseValue() {
		return false
	}

	// Save the current token info for potential map detection.
	firstValueType := p.current.Typ
	firstValueVal := p.current.Val

	// Check if this is a map, the next token would be `:`.
	if p.advance() && p.current.Typ == TokenPairSeparator {
		// It's a map, parse as a map starting with the saved first key.
		return p.parseMapFrom(firstValueType, firstValueVal)
	}

	// It's a regular list.
	p.emit(ListStartEvent{})
	p.emit(ValueEvent{Type: firstValueType, Value: firstValueVal})

	for p.hasToken && p.current.Typ == TokenListSeparator {
		if !p.advance() || !p.parseValue() {
			return false
		}
		p.emit(ValueEvent{Type: p.current.Typ, Value: p.current.Val})

		// Advance to the next token to check for more separators.
		if !p.advance() {
			break
		}
	}

	p.emit(ListEndEvent{})
	return true
}

// parseMapFrom parses a map starting from a known first key.
func (p *Parser) parseMapFrom(keyType TokenType, keyVal string) bool {
	p.emit(MapStartEvent{})
	p.emit(MapKeyEvent{Type: keyType, Value: keyVal})

	// We're already at the `:` token.
	if !p.expect(TokenPairSeparator) {
		return false
	}

	// Advance to the next token for the value.
	if !p.advance() || !p.parseValue() {
		return false
	}
	p.emit(ValueEvent{Type: p.current.Typ, Value: p.current.Val})

	// Parse the remaining key-value pairs.
	for p.advance() && p.current.Typ == TokenListSeparator {
		if !p.advance() || !p.parseValue() {
			return false
		}

		p.emit(MapKeyEvent{Type: p.current.Typ, Value: p.current.Val})

		if !p.advance() || !p.expect(TokenPairSeparator) {
			return false
		}

		if !p.advance() || !p.parseValue() {
			return false
		}
		p.emit(ValueEvent{Type: p.current.Typ, Value: p.current.Val})
	}

	p.emit(MapEndEvent{})
	return true
}

// parseValue parses a single value
func (p *Parser) parseValue() bool {
	switch p.current.Typ {
	case TokenIdentifier, TokenNumber, TokenString, TokenTrue, TokenFalse, TokenNil:
		return true
	default:
		return p.errorf("expected value, got %s", p.current.Typ)
	}
}

// Parse returns an iterator that yields parse events
func Parse(tokens iter.Seq[Token]) iter.Seq[ParserEvent] {
	return func(yield func(ParserEvent) bool) {
		next, stop := iter.Pull(tokens)
		defer stop()

		p := &Parser{
			yield: yield,
			next:  next,
		}

		p.parseFieldList()
	}
}
