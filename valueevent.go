package plainfields

import (
	"fmt"
	"strconv"
)

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
	return strconv.ParseInt(e.Value, 10, 64)
}

// ToUint returns the value as an unsigned integer.
func (e ValueEvent) ToUint() (uint64, error) {
	if e.Type != TokenNumber {
		return 0, fmt.Errorf("cannot convert %s to uint", e.Type)
	}
	return strconv.ParseUint(e.Value, 10, 64)
}

// ToFloat returns the value as a floating point number.
func (e ValueEvent) ToFloat() (float64, error) {
	if e.Type != TokenNumber {
		return 0, fmt.Errorf("cannot convert %s to float", e.Type)
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
