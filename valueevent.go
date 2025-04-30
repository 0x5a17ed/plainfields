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

// IsNil checks if the value is nil.
func (e ValueEvent) IsNil() bool {
	return e.Type == TokenNil
}

// ToString returns the value as a string.
func (e ValueEvent) ToString() (string, error) {
	switch e.Type {
	case TokenString:
		return strconv.Unquote(e.Value)
	case TokenIdentifier, TokenNumber:
		return e.Value, nil
	default:
		return "", fmt.Errorf("cannot convert %s to string", e.Type)
	}
}

// ToInt returns the value as an integer.
func (e ValueEvent) ToInt() (int64, error) {
	if e.Type != TokenNumber {
		return 0, fmt.Errorf("cannot convert %s to int", e.Type)
	}
	return strconv.ParseInt(e.Value, 0, 64)
}

// ToUint returns the value as an unsigned integer.
func (e ValueEvent) ToUint() (uint64, error) {
	if e.Type != TokenNumber {
		return 0, fmt.Errorf("cannot convert %s to uint", e.Type)
	}
	return strconv.ParseUint(e.Value, 0, 64)
}

// ToFloat returns the value as a floating point number.
func (e ValueEvent) ToFloat() (float64, error) {
	if e.Type != TokenNumber {
		return 0, fmt.Errorf("cannot convert %s to float", e.Type)
	}

	// Remove underscores used as separators.
	s := strings.ReplaceAll(e.Value, "_", "")

	// Handle different number bases.
	if len(s) > 3 && s[0] == '0' {
		switch s[1] {
		case 'x', 'X':
			// Check for hex float with 'p' exponent
			if strings.ContainsAny(s, "pP") {
				return strconv.ParseFloat(e.Value, 64)
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

	return strconv.ParseFloat(e.Value, 64)
}

// ToBool returns the value as a boolean.
func (e ValueEvent) ToBool() (bool, error) {
	switch e.Type {
	case TokenTrue:
		return true, nil
	case TokenFalse:
		return false, nil
	default:
		return false, fmt.Errorf("cannot convert %s to bool", e.Type)
	}
}
