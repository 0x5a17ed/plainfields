package kaval

import (
	"fmt"
	"iter"
	"slices"
	"strconv"
)

// ParseOptions holds options for parsing.
type ParseOptions struct {
	// AllowOrdered allows ordered values without a key.
	AllowOrdered bool
}

// ParseDefaults returns the default parsing options.
func ParseDefaults() ParseOptions {
	return ParseOptions{
		AllowOrdered: true,
	}
}

// ParserEvent represents different events that can occur during parsing.
type ParserEvent interface {
	isParserEvent()
}

type (
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
		Value
	}

	// ValueEvent represents any value found.
	ValueEvent struct {
		Value
	}

	// ErrorEvent represents an error during parsing.
	ErrorEvent struct {
		Pos Position
		Msg string
	}
)

// Error implements the error interface for ErrorEvent.
func (e ErrorEvent) Error() string {
	return fmt.Sprintf("Error at %s: %s", e.Pos, e.Msg)
}

// Unwrap returns the internal value of the event.
func (e ValueEvent) Unwrap() Value {
	return e.Value
}

// Unwrap returns the internal value of the event.
func (e MapKeyEvent) Unwrap() Value {
	return e.Value
}

func (ValueEvent) isParserEvent()     {}
func (ListStartEvent) isParserEvent() {}
func (ListEndEvent) isParserEvent()   {}
func (MapStartEvent) isParserEvent()  {}
func (MapEndEvent) isParserEvent()    {}
func (MapKeyEvent) isParserEvent()    {}
func (ErrorEvent) isParserEvent()     {}

type parserState int

const (
	startState parserState = iota
	orderedState
	labeledState
	eofState
)

// Parser holds the state for parsing.
type Parser struct {
	config ParseOptions
	yield  func(ParserEvent) bool
	next   func() (Token, bool)

	done     bool
	peeked   *Token
	current  Token
	hasToken bool

	state parserState
}

// emit sends an event through yield.
func (p *Parser) emit(event ParserEvent) bool {
	p.done = p.done || !p.yield(event)
	return !p.done
}

// errorf sends an error event through yield.
func (p *Parser) errorf(format string, args ...any) bool {
	p.emit(ErrorEvent{
		Pos: p.current.Pos,
		Msg: fmt.Sprintf(format, args...),
	})
	return false
}

// peek looks at the next token without consuming it.
func (p *Parser) peek() *Token {
	// If we have not peeked yet, get the next token.
	if p.peeked == nil {
		// Fetch the next token and store it.
		if tok, ok := p.next(); ok {
			p.peeked = &tok
		}
	}
	return p.peeked
}

// isNext checks if the next token is of the expected type.
func (p *Parser) isNext(typ ...TokenType) bool {
	t := p.peek()
	return t != nil && slices.Contains(typ, t.Typ)
}

// advance advances to the next token.
func (p *Parser) advance() bool {
	if p.peeked != nil {
		// If we have a peeked token, consume it.
		p.current = *p.peeked
		p.peeked = nil
		p.hasToken = true

	} else if tok, ok := p.next(); ok {
		// If we have no peeked token, get the next token.
		p.hasToken = true
		p.current = tok

	} else {
		p.hasToken = false
	}

	return p.hasToken
}

// isToken checks if the current token is of the expected type
func (p *Parser) isToken(typ TokenType) bool {
	if !p.hasToken {
		return p.errorf("unexpected end of input")
	}
	if p.current.Typ != typ {
		return p.errorf("expected %s, got %s", typ, p.current.Typ)
	}
	return true
}

// toValue converts the current token to a Value.
func (p *Parser) toValue() Value {
	return valueFromToken(p.current)
}

// updateState updates the parser state.
func (p *Parser) updateState(newState parserState) {
	switch {
	case p.state == startState && newState == orderedState:
		// If we are starting a new ordered section, emit the list start event.
		p.emit(ListStartEvent{})
	case p.state == startState && newState == labeledState:
		// If we are starting a new labeled section, emit the map start event.
		p.emit(MapStartEvent{})
	case p.state == orderedState && newState == labeledState:
		// Transition from ordered state to the labeled section state.
		p.emit(ListEndEvent{})
		p.emit(MapStartEvent{})
	case newState == eofState:
		// If we are at the end of the file, emit the end event.
		if p.state == orderedState {
			p.emit(ListEndEvent{})
		} else if p.state == labeledState {
			p.emit(MapEndEvent{})
		}
	}

	p.state = newState
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
	p.updateState(eofState)
}

// parseField parses a single field
func (p *Parser) parseField() bool {
	switch p.current.Typ {
	case TokenBooleanPrefix:
		p.updateState(labeledState)
		return p.parseBooleanPrefix()

	case TokenIdentifier:
		// Check if this is a labeled field assignment.
		if p.isNext(TokenAssign) {
			p.updateState(labeledState)
			return p.parseAssignment()
		}

		// If it's not an assignment, treat it as an ordered value.
		fallthrough

	case TokenFieldSeparator, TokenString, TokenNumber, TokenTrue, TokenFalse, TokenNil:
		if !p.config.AllowOrdered || p.state > orderedState {
			return p.errorf("ordered value not allowed here")
		}
		p.updateState(orderedState)

		if p.current.Typ == TokenFieldSeparator {
			p.emit(ValueEvent{ZeroValue{}})
			return true
		} else {
			return p.parseValueContent()
		}

	default:
		return p.errorf("expected field prefix, identifier, or value, got %s", p.current.Typ)
	}
}

// parseBooleanPrefix parses a field with a prefix (^ or !)
func (p *Parser) parseBooleanPrefix() bool {
	prefix := p.current.Val // '^' or '!'
	if !p.advance() || !p.isToken(TokenIdentifier) {
		return false
	}

	// Emit as a boolean assignment.
	p.emit(MapKeyEvent{p.toValue()})
	p.emit(ValueEvent{BooleanValue{strconv.FormatBool(prefix == "^")}})

	return p.advance()
}

// parseAssignment handles key=value fields
func (p *Parser) parseAssignment() bool {
	p.emit(MapKeyEvent{p.toValue()})

	p.advance() // Consume the `=` token.

	// If the next token is a field separator or EOF, it's an empty assignment.
	if p.isNext(TokenFieldSeparator, TokenEOF) {
		p.advance()                     // Consume the assignment token.
		p.emit(ValueEvent{ZeroValue{}}) // Emit a zero value.
		return true
	}

	// We're already past the assign token.
	if !p.advance() || !p.parseValueContent() {
		return false
	}

	return true
}

// parseValueContent parses a list of values, detecting if it's a map.
func (p *Parser) parseValueContent() bool {
	// Handle boolean prefix notation.
	if p.current.Typ == TokenBooleanPrefix {
		return p.parseDictValue()
	}

	if !p.isValue() {
		return false
	}

	switch {
	case p.isNext(TokenPairSeparator):
		// It's a map, parse as a map starting with the first key.
		return p.parseDictValue()

	case p.isNext(TokenListSeparator):
		// It's a list, parse as a list starting with the first value.
		return p.parseListValue()

	default:
		// If we don't have a list or map, just emit a single value.
		p.emit(ValueEvent{p.toValue()})
		p.advance() // Consume the value.
		return true
	}
}

// parseListValue parses a list starting from a known first value.
func (p *Parser) parseListValue() bool {
	// It's a regular list.
	p.emit(ListStartEvent{})
	p.emit(ValueEvent{p.toValue()})

	// Advance to the next token for the list separator.
	p.advance()

	for p.hasToken && p.current.Typ == TokenListSeparator {
		if !p.advance() || !p.isValue() {
			return false
		}
		p.emit(ValueEvent{p.toValue()})

		// Advance to the next token to check for more separators.
		p.advance()
	}

	p.emit(ListEndEvent{})
	return true
}

// parseDictValue parses a map starting from a known first key.
func (p *Parser) parseDictValue() bool {
	p.emit(MapStartEvent{})

	if !p.parseDictEntry() {
		return false
	}

	// ParseTokens the remaining key-value pairs.
	for p.current.Typ == TokenListSeparator {
		if !p.advance() || !p.parseDictEntry() {
			return false
		}
	}

	p.emit(MapEndEvent{})
	return true
}

// parseDictEntry parses a single key-value pair in a map.
func (p *Parser) parseDictEntry() bool {
	if p.current.Typ == TokenBooleanPrefix {
		return p.parseBooleanPrefix()
	}

	return p.parseDictPair()
}

// parseDictPair parses a key-value pair in a map.
func (p *Parser) parseDictPair() bool {
	// The current parser position should be a key.
	if !p.isValue() {
		return false
	}
	p.emit(MapKeyEvent{p.toValue()})

	// Parse the colon between key and value.
	if !p.advance() || !p.isToken(TokenPairSeparator) {
		return false
	}

	// Parse the value.
	if !p.advance() || !p.isValue() {
		return false
	}

	return p.emit(ValueEvent{p.toValue()}) && p.advance()
}

// isValue parses a single value
func (p *Parser) isValue() bool {
	switch p.current.Typ {
	case TokenIdentifier, TokenNumber, TokenString, TokenTrue, TokenFalse, TokenNil:
		return true
	default:
		return p.errorf("expected value, got %s", p.current.Typ)
	}
}

// ParseTokens returns an iterator that yields parse events
func ParseTokens(tokens iter.Seq[Token], opts ...ParseOptions) iter.Seq[ParserEvent] {
	opt := ParseDefaults()
	if len(opts) > 0 {
		opt = opts[0]
	}

	return func(yield func(ParserEvent) bool) {
		next, stop := iter.Pull(tokens)
		defer stop()

		p := &Parser{
			config: opt,
			yield:  yield,
			next:   next,
		}

		p.parseFieldList()
	}
}

// Parse parses the input string and returns an iterator of ParserEvent.
func Parse(input string, opts ...ParseOptions) iter.Seq[ParserEvent] {
	return ParseTokens(Lex(input), opts...)
}
