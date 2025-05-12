package plainfields

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ValueType int

const (
	InvalidValueType ValueType = iota
	ZeroValueType
	NilValueType
	BooleanValueType
	NumberValueType
	StringValueType
	IdentifierValueType
)

// GoString returns the Go string representation of the ValueType.
func (vt ValueType) GoString() string {
	switch vt {
	case InvalidValueType:
		return "InvalidValueType"
	case ZeroValueType:
		return "ZeroValueType"
	case NilValueType:
		return "NilValueType"
	case BooleanValueType:
		return "BooleanValueType"
	case NumberValueType:
		return "NumberValueType"
	case StringValueType:
		return "StringValueType"
	case IdentifierValueType:
		return "IdentifierValueType"
	default:
		return fmt.Sprintf("ValueType(%d)", vt)
	}
}

// String returns the string representation of the ValueType.
func (vt ValueType) String() string {
	switch vt {
	case InvalidValueType:
		return "invalid"
	case ZeroValueType:
		return "zero"
	case NilValueType:
		return "nil"
	case BooleanValueType:
		return "boolean"
	case NumberValueType:
		return "number"
	case StringValueType:
		return "string"
	case IdentifierValueType:
		return "identifier"
	default:
		return fmt.Sprintf("ValueType(%d)", vt)
	}
}

// Value represents a plainfields value.
type Value interface {
	fmt.Stringer
	Type() ValueType
	Raw() string
}

type NilValue struct{}

func (v NilValue) String() string  { return fmt.Sprintf("%s (%s)", v.Raw(), v.Type()) }
func (v NilValue) Type() ValueType { return NilValueType }
func (v NilValue) Raw() string     { return "nil" }
func (v NilValue) IsZero() bool    { return true }

func IsNil(v Value) bool { return v.Type() == NilValueType }

type ZeroValue struct{}

func (v ZeroValue) String() string  { return fmt.Sprintf("%s (%s)", v.Raw(), v.Type()) }
func (v ZeroValue) Type() ValueType { return ZeroValueType }
func (v ZeroValue) Raw() string     { return "zero" }
func (v ZeroValue) IsZero() bool    { return true }

type IdentifierValue struct{ raw string }

func (v IdentifierValue) String() string  { return fmt.Sprintf("%s (%s)", v.Raw(), v.Type()) }
func (v IdentifierValue) Type() ValueType { return IdentifierValueType }
func (v IdentifierValue) Raw() string     { return v.raw }
func (v IdentifierValue) ToString() (string, error) {
	return v.raw, nil
}

type StringValue struct {
	raw string
}

func (v StringValue) String() string  { return fmt.Sprintf("%s (%s)", v.Raw(), v.Type()) }
func (v StringValue) Type() ValueType { return StringValueType }
func (v StringValue) Raw() string     { return v.raw }
func (v StringValue) IsZero() bool    { return len(v.raw) < 2 }
func (v StringValue) ToString() (string, error) {
	return strconv.Unquote(v.raw)
}

type NumberValue struct{ raw string }

func (v NumberValue) String() string          { return fmt.Sprintf("%s (%s)", v.Raw(), v.Type()) }
func (v NumberValue) Type() ValueType         { return NumberValueType }
func (v NumberValue) Raw() string             { return v.raw }
func (v NumberValue) IsSigned() bool          { return strings.HasPrefix(v.raw, "-") }
func (v NumberValue) IsFloat() bool           { return strings.ContainsAny(v.raw, ".eEpP") }
func (v NumberValue) ToUint() (uint64, error) { return strconv.ParseUint(v.raw, 0, 64) }
func (v NumberValue) ToInt() (int64, error)   { return strconv.ParseInt(v.raw, 0, 64) }
func (v NumberValue) IsZero() bool {
	n, err := v.ToFloat()
	return err == nil && n == 0
}
func (v NumberValue) ToFloat() (float64, error) {
	s := strings.ReplaceAll(v.raw, "_", "")
	if len(s) > 2 && s[0] == '0' {
		switch s[1] {
		case 'x', 'X':
			if strings.ContainsAny(s, "pP") {
				return strconv.ParseFloat(s, 64)
			}
			if strings.Contains(s, ".") {
				return parseHexFloat(s)
			}
			v, err := strconv.ParseUint(s[2:], 16, 64)
			return float64(v), err
		case 'o', 'O':
			v, err := strconv.ParseUint(s[2:], 8, 64)
			return float64(v), err
		case 'b', 'B':
			v, err := strconv.ParseUint(s[2:], 2, 64)
			return float64(v), err
		}
	}
	return strconv.ParseFloat(s, 64)
}

func parseHexFloat(s string) (float64, error) {
	// Split into mantissa and exponent
	mantissaStr := s[2:]

	dotIdx := strings.IndexByte(mantissaStr, '.')
	intPart := mantissaStr[:dotIdx]
	fracPart := mantissaStr[dotIdx+1:]

	var mantissa float64

	// Parse integer part
	if intPart != "" {
		v, err := strconv.ParseUint(intPart, 16, 64)
		if err != nil {
			return 0, err
		}
		mantissa = float64(v)
	}

	// Parse fractional part
	if fracPart != "" {
		v, err := strconv.ParseUint(fracPart, 16, 64)
		if err != nil {
			return 0, err
		}
		mantissa += float64(v) / math.Pow(16, float64(len(fracPart)))
	}

	return mantissa, nil
}

type BooleanValue struct{ raw string }

func (v BooleanValue) String() string  { return fmt.Sprintf("%s (%s)", v.Raw(), v.Type()) }
func (v BooleanValue) IsZero() bool    { return v.raw == "false" }
func (v BooleanValue) Type() ValueType { return BooleanValueType }
func (v BooleanValue) Raw() string     { return v.raw }
func (v BooleanValue) ToBool() (bool, error) {
	switch v.raw {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean: %s", v.raw)
	}
}

// valueFromToken converts a token to a Value.
func valueFromToken(token Token) Value {
	switch token.Typ {
	case TokenString:
		return StringValue{raw: token.Val}
	case TokenNumber:
		return NumberValue{raw: token.Val}
	case TokenIdentifier:
		return IdentifierValue{raw: token.Val}
	case TokenFalse, TokenTrue:
		return BooleanValue{raw: token.Val}
	case TokenNil:
		return NilValue{}
	default:
		return nil
	}
}

// As attempts to convert a value to a given type.
func As[T any](v any) (val T, ok bool) {
	if val, ok = v.(T); ok {
		return val, true
	}

	if conv, ok := v.(interface{ Unwrap() Value }); ok {
		return As[T](conv.Unwrap())
	}

	var zero T
	return zero, false
}

// IsZero checks if a Value is a zero value.
func IsZero(v Value) bool {
	if check, ok := As[interface{ IsZero() bool }](v); ok {
		return check.IsZero()
	}

	return false
}

// ToString attempts to convert a Value to a string value.
func ToString(v Value) (string, error) {
	if conv, ok := As[interface{ ToString() (string, error) }](v); ok {
		return conv.ToString()
	}

	return "", fmt.Errorf("value of type %s is not string-convertible", v.Type())
}

// ToFloat attempts to convert a Value to a float value.
func ToFloat(v Value) (float64, error) {
	if conv, ok := As[interface{ ToFloat() (float64, error) }](v); ok {
		return conv.ToFloat()
	}
	return 0, fmt.Errorf("value of type %s is not float-convertible", v.Type())
}

// ToInt attempts to convert a Value to an int64 value.
func ToInt(v Value) (int64, error) {
	if conv, ok := As[interface{ ToInt() (int64, error) }](v); ok {
		return conv.ToInt()
	}

	return 0, fmt.Errorf("value of type %s is not int-convertible", v.Type())
}

// ToUint attempts to convert a Value to an uint64 value.
func ToUint(v Value) (uint64, error) {
	if conv, ok := As[interface{ ToUint() (uint64, error) }](v); ok {
		return conv.ToUint()
	}

	return 0, fmt.Errorf("value of type %s is not uint-convertible", v.Type())
}

// ToBool attempts to convert a Value to a boolean value.
func ToBool(v Value) (bool, error) {
	if conv, ok := As[interface{ ToBool() (bool, error) }](v); ok {
		return conv.ToBool()
	}
	return false, fmt.Errorf("value of type %s is not bool-convertible", v.Type())
}
