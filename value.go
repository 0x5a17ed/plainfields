package plainfields

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

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
type Value struct {
	Type  ValueType
	Value string
}

var ZeroValue = Value{Type: ZeroValueType, Value: ""}

// IsZero checks if the value is zero.
func (v Value) IsZero() bool {
	return v.Type == ZeroValueType
}

// IsNil checks if the value is nil.
func (v Value) IsNil() bool {
	return v.Type == NilValueType
}

// ToString returns the value as a string.
func (v Value) ToString() (string, error) {
	switch v.Type {
	case StringValueType:
		return strconv.Unquote(v.Value)
	case IdentifierValueType, NumberValueType:
		return v.Value, nil
	default:
		return "", fmt.Errorf("cannot convert %s to string", v.Type)
	}
}

// ToInt returns the value as an integer.
func (v Value) ToInt() (int64, error) {
	if v.Type != NumberValueType {
		return 0, fmt.Errorf("cannot convert %s to int", v.Type)
	}
	return strconv.ParseInt(v.Value, 0, 64)
}

// ToUint returns the value as an unsigned integer.
func (v Value) ToUint() (uint64, error) {
	if v.Type != NumberValueType {
		return 0, fmt.Errorf("cannot convert %s to uint", v.Type)
	}
	return strconv.ParseUint(v.Value, 0, 64)
}

// ToFloat returns the value as a floating point number.
func (v Value) ToFloat() (float64, error) {
	if v.Type != NumberValueType {
		return 0, fmt.Errorf("cannot convert %s to float", v.Type)
	}

	// Remove underscores used as separators.
	s := strings.ReplaceAll(v.Value, "_", "")

	// Handle different number bases.
	if len(s) > 3 && s[0] == '0' {
		switch s[1] {
		case 'x', 'X':
			// Check for hex float with 'p' exponent
			if strings.ContainsAny(s, "pP") {
				return strconv.ParseFloat(v.Value, 64)
			} else if strings.Contains(s, ".") {
				return parseHexFloat(s)
			}

			// Regular hex.
			v, err := strconv.ParseUint(s[2:], 16, 64)
			if err != nil {
				return 0, err
			}
			return float64(v), nil

		case 'o', 'O':
			// Octal.
			v, err := strconv.ParseUint(s[2:], 8, 64)
			if err != nil {
				return 0, err
			}
			return float64(v), nil

		case 'b', 'B':
			// Binary.
			v, err := strconv.ParseUint(s[2:], 2, 64)
			if err != nil {
				return 0, err
			}
			return float64(v), nil
		}
	}

	return strconv.ParseFloat(v.Value, 64)
}

// ToBool returns the value as a boolean.
func (v Value) ToBool() (bool, error) {
	if v.Type != BooleanValueType {
		return false, fmt.Errorf("cannot convert %s to bool", v.Type)
	}

	return v.Value == "true", nil
}

// valueFromToken converts a token to a Value.
func valueFromToken(token Token) Value {
	switch token.Typ {
	case TokenString:
		return Value{Type: StringValueType, Value: token.Val}
	case TokenNumber:
		return Value{Type: NumberValueType, Value: token.Val}
	case TokenIdentifier:
		return Value{Type: IdentifierValueType, Value: token.Val}
	case TokenFalse, TokenTrue:
		return Value{Type: BooleanValueType, Value: token.Val}
	case TokenNil:
		return Value{Type: NilValueType, Value: "nil"}

	default:
		return Value{}
	}
}
