package plainfields

import (
	"fmt"
	"iter"
	"slices"
)

// ParseOptions holds options for parsing.
type ParseOptions struct {
	// AllowPositional allows positional values without a key.
	AllowPositional bool
}

// ParseOptionsDefaults returns the default parsing options.
func ParseOptionsDefaults() ParseOptions {
	return ParseOptions{
		AllowPositional: true,
	}
}

// ParserEvent represents different events that can occur during parsing.
type ParserEvent interface {
	isParserEvent()
}

type (
	// OrderedFieldStartEvent represents a positional field value assignment without a key.
	OrderedFieldStartEvent struct {
		Index int
	}

	// LabeledFieldStartEvent represents the beginning of a field value assignment with a key.
	LabeledFieldStartEvent struct {
		Name string
	}

	// FieldEndEvent represents the end of a field value assignment.
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

// Error implements the error interface for ErrorEvent.
func (e ErrorEvent) Error() string {
	return fmt.Sprintf("Error at %s: %s", e.Pos, e.Msg)
}

func (OrderedFieldStartEvent) isParserEvent() {}
func (LabeledFieldStartEvent) isParserEvent() {}
func (FieldEndEvent) isParserEvent()          {}
func (ValueEvent) isParserEvent()             {}
func (ListStartEvent) isParserEvent()         {}
func (ListEndEvent) isParserEvent()           {}
func (MapStartEvent) isParserEvent()          {}
func (MapEndEvent) isParserEvent()            {}
func (MapKeyEvent) isParserEvent()            {}
func (ErrorEvent) isParserEvent()             {}

// Parser holds the state for parsing.
type Parser struct {
	yield func(ParserEvent) bool
	next  func() (Token, bool)

	done     bool
	peeked   *Token
	current  Token
	hasToken bool

	fieldIndex     int
	assignmentMode bool
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
		p.assignmentMode = true
		return p.parseFieldPrefix()

	case TokenIdentifier:
		// Check if this is a labeled field assignment.
		if p.isNext(TokenAssign) {
			p.assignmentMode = true
			return p.parseAssignment()
		}

		// If it's not an assignment, treat it as a positional value.
		fallthrough

	case TokenString, TokenNumber, TokenTrue, TokenFalse, TokenNil:
		if p.assignmentMode {
			return p.errorf("positional value not allowed here")
		}

		p.emit(OrderedFieldStartEvent{p.fieldIndex})
		p.fieldIndex++
		p.emit(ValueEvent{Type: p.current.Typ, Value: p.current.Val})
		p.emit(FieldEndEvent{})
		return p.advance()

	default:
		return p.errorf("expected field prefix, identifier, or value, got %s", p.current.Typ)
	}
}

// parseFieldPrefix parses a field with a prefix (^ or !)
func (p *Parser) parseFieldPrefix() bool {
	prefix := p.current.Val // '^' or '!'
	if !p.advance() || !p.expect(TokenIdentifier) {
		return false
	}
	name := p.current.Val

	// Emit as a boolean assignment.
	p.emit(LabeledFieldStartEvent{name})
	if prefix == "^" {
		p.emit(ValueEvent{TokenTrue, "true"})
	} else { // `!`
		p.emit(ValueEvent{TokenFalse, "false"})
	}
	p.emit(FieldEndEvent{})
	return p.advance()
}

// parseAssignment handles key=value fields
func (p *Parser) parseAssignment() bool {
	p.emit(LabeledFieldStartEvent{Name: p.current.Val})

	p.advance() // Consume the `=` token.

	// If the next token is a field separator or EOF, it's an empty assignment.
	if p.isNext(TokenFieldSeparator, TokenEOF) {
		p.advance() // Consume the assignment token.

		p.emit(FieldEndEvent{})
		return true
	}

	// We're already past the assign token.
	if !p.advance() || !p.parseValueContent() {
		return false
	}

	p.emit(FieldEndEvent{})
	return true
}

// parseValueContent parses a list of values, detecting if it's a map.
func (p *Parser) parseValueContent() bool {
	if !p.parseValue() {
		return false
	}

	switch {
	case p.isNext(TokenPairSeparator):
		// It's a map, parse as a map starting with the first key.
		return p.parseMapFrom()

	case p.isNext(TokenListSeparator):
		// It's a list, parse as a list starting with the first value.
		return p.parseListFrom()

	default:
		// If we don't have a list or map, just emit a single value.
		p.emit(ValueEvent{Type: p.current.Typ, Value: p.current.Val})
		p.advance() // Consume the value.
		return true
	}
}

// parseListFrom parses a list starting from a known first value.
func (p *Parser) parseListFrom() bool {
	// It's a regular list.
	p.emit(ListStartEvent{})
	p.emit(ValueEvent{Type: p.current.Typ, Value: p.current.Val})

	// Advance to the next token for the list separator.
	p.advance()

	for p.hasToken && p.current.Typ == TokenListSeparator {
		if !p.advance() || !p.parseValue() {
			return false
		}
		p.emit(ValueEvent{Type: p.current.Typ, Value: p.current.Val})

		// Advance to the next token to check for more separators.
		p.advance()
	}

	p.emit(ListEndEvent{})
	return true
}

// parseMapFrom parses a map starting from a known first key.
func (p *Parser) parseMapFrom() bool {
	p.emit(MapStartEvent{})

	if !p.parseKeyValuePair() {
		return false
	}

	// ParseTokens the remaining key-value pairs.
	for p.advance() && p.current.Typ == TokenListSeparator {
		if !p.advance() || !p.parseValue() {
			return false
		}

		if !p.parseKeyValuePair() {
			return false
		}
	}

	p.emit(MapEndEvent{})
	return true
}

// parseKeyValuePair parses a key-value pair in a map.
func (p *Parser) parseKeyValuePair() bool {
	p.emit(MapKeyEvent{Type: p.current.Typ, Value: p.current.Val})

	// Parse the colon between key and value.
	if !p.advance() || !p.expect(TokenPairSeparator) {
		return false
	}

	// Parse the value.
	if !p.advance() || !p.parseValue() {
		return false
	}

	return p.emit(ValueEvent{Type: p.current.Typ, Value: p.current.Val})
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

// ParseTokens returns an iterator that yields parse events
func ParseTokens(tokens iter.Seq[Token], opts ...ParseOptions) iter.Seq[ParserEvent] {
	opt := ParseOptionsDefaults()
	if len(opts) > 0 {
		opt = opts[0]
	}

	return func(yield func(ParserEvent) bool) {
		next, stop := iter.Pull(tokens)
		defer stop()

		p := &Parser{
			yield: yield,
			next:  next,
		}

		// Set assignment mode if positional values are not allowed.
		if !opt.AllowPositional {
			p.assignmentMode = true
		}

		p.parseFieldList()
	}
}

// Parse parses the input string and returns an iterator of ParserEvent.
func Parse(input string, opts ...ParseOptions) iter.Seq[ParserEvent] {
	return ParseTokens(Lex(input), opts...)
}
