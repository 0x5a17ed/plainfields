package kaval

import (
	"fmt"
	"testing"
)

func TestValueType_GoString(t *testing.T) {
	tests := []struct {
		vt   ValueType
		want string
	}{
		{InvalidValueType, "InvalidValueType"},
		{ZeroValueType, "ZeroValueType"},
		{NilValueType, "NilValueType"},
		{BooleanValueType, "BooleanValueType"},
		{NumberValueType, "NumberValueType"},
		{StringValueType, "StringValueType"},
		{IdentifierValueType, "IdentifierValueType"},
		{ValueType(999), "ValueType(999)"}, // unknown case
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.vt.GoString()
			if got != tt.want {
				t.Errorf("GoString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// collectEvent collects a single event from the parser.
func collectEvent[T any](input string, options ...ParseOptions) (T, error) {
	var zero T

	for event := range Parse(input, options...) {
		if err, isError := event.(ErrorEvent); isError {
			return zero, &err
		} else if ev, ok := event.(T); ok {
			return ev, nil
		}
	}

	return zero, fmt.Errorf("no event found")
}

func p[T any](v T) *T {
	return &v
}

// testConversion is a generic helper for testing value conversions
func testConversion[T comparable](t *testing.T, name string, want *T, v Value, fn func(v Value) (T, error)) {
	if got, err := fn(v); want != nil {
		// We isToken success with a specific value
		if err != nil {
			t.Fatalf("%s() error = %v", name, err)
		}
		if got != *want {
			t.Errorf("%s() = %v, want %v", name, got, *want)
		}
	} else if err == nil {
		t.Fatalf("%s() expected error, got nil", name)
	}
}

func TestValue_Conversions(t *testing.T) {
	tests := []struct {
		input string

		wantRaw    *string
		wantString *string
		wantInt    *int64
		wantUint   *uint64
		wantFloat  *float64
		wantBool   *bool
		wantNil    bool
		wantZero   bool
	}{
		{input: `,`, wantRaw: p("zero"), wantZero: true},

		{input: `field`, wantString: p("field")},
		{input: `"hello"`, wantString: p("hello")},
		{input: `-42`, wantInt: p(int64(-42)), wantFloat: p(float64(-42))},
		{input: `23`, wantInt: p(int64(23)), wantUint: p(uint64(23)), wantFloat: p(float64(23))},
		{input: `3.14`, wantFloat: p(float64(3.14))},

		{input: "0x23", wantInt: p(int64(0x23)), wantUint: p(uint64(0x23)), wantFloat: p(float64(0x23))},
		{input: "0x23.1", wantFloat: p(float64(35.0625))},
		{input: "0x1p1", wantFloat: p(float64(2))},
		{input: "0x1.8p1", wantFloat: p(float64(3.0))},
		{input: "0x1.8p-1", wantFloat: p(float64(0.75))},
		{input: "0x1.8p+1", wantFloat: p(float64(3.0))},

		{input: "0o644", wantInt: p(int64(420)), wantUint: p(uint64(420)), wantFloat: p(float64(420))},
		{input: "0b01100101", wantInt: p(int64(101)), wantUint: p(uint64(101)), wantFloat: p(float64(101))},

		{input: `true`, wantBool: p(true)},
		{input: `false`, wantBool: p(false), wantZero: true},
		{input: `nil`, wantNil: true, wantZero: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ev, err := collectEvent[ValueEvent](tt.input)
			if err != nil {
				t.Fatalf("collectEvent() error = %v", err)
			}

			testConversion(t, "ToString", tt.wantString, ev.Value, ToString)
			testConversion(t, "ToInt", tt.wantInt, ev.Value, ToInt)
			testConversion(t, "ToUint", tt.wantUint, ev.Value, ToUint)
			testConversion(t, "ToFloat", tt.wantFloat, ev.Value, ToFloat)
			testConversion(t, "ToBool", tt.wantBool, ev.Value, ToBool)

			if tt.wantRaw == nil {
				tt.wantRaw = &tt.input
			}

			if ev.Raw() != *tt.wantRaw {
				t.Errorf("Raw() = %q, want %q", ev.Value.Raw(), tt.input)
			}

			if got := IsNil(ev); got != tt.wantNil {
				t.Errorf("IsNil() = %t, want %t", got, tt.wantNil)
			}

			if got := IsZero(ev); got != tt.wantZero {
				t.Errorf("IsZero() = %t, want %t", got, tt.wantZero)
			}
		})
	}
}

func TestNumberValue(t *testing.T) {
	tt := []struct {
		inp      string
		isFloat  bool
		isSigned bool
	}{
		{inp: "-1.5", isFloat: true, isSigned: true},
		{inp: "-0x1.8p1", isFloat: true, isSigned: true},
		{inp: "-0x1.8p-1", isFloat: true, isSigned: true},
		{inp: "-0x1.8p+1", isFloat: true, isSigned: true},
		{inp: "-0x1.8p0", isFloat: true, isSigned: true},

		{inp: "0x1.8p1", isFloat: true, isSigned: false},
		{inp: "0x1.8p-1", isFloat: true, isSigned: false},
		{inp: "0x1.8p+1", isFloat: true, isSigned: false},
		{inp: "0x1.8p0", isFloat: true, isSigned: false},

		{inp: "3.14", isFloat: true, isSigned: false},
		{inp: "3.14e10", isFloat: true, isSigned: false},
		{inp: "3.14e-10", isFloat: true, isSigned: false},
		{inp: "1.5e10", isFloat: true, isSigned: false},
		{inp: "1e10", isFloat: true, isSigned: false},

		{inp: "1", isFloat: false, isSigned: false},
		{inp: "-1", isFloat: false, isSigned: true},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.inp, func(t *testing.T) {
			v := NumberValue{tc.inp}

			t.Run("IsFloat", func(t *testing.T) {
				if v.IsFloat() != tc.isFloat {
					t.Errorf("IsFloat() = %t, want %t", v.IsFloat(), tc.isFloat)
				}
			})

			t.Run("IsSigned", func(t *testing.T) {
				if v.IsSigned() != tc.isSigned {
					t.Errorf("IsSigned() = %t, want %t", v.IsSigned(), tc.isSigned)
				}
			})
		})
	}

}

func TestNumberValue_ToFloatError(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"0x1foo"},
		{"0x1foo.bar"},
		{"0x1.bar"},
		{"0octal"},
		{"0binary"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ev := newValue(NumberValueType, tt.input)

			_, err := ToFloat(ev)
			if err == nil {
				t.Fatalf("ToFloat() error = %v", err)
			}
		})
	}
}

func TestValueStringer(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{"NilValue", NilValue{}, "nil (nil)"},
		{"ZeroValue", ZeroValue{}, "zero (zero)"},
		{"BooleanValue true", BooleanValue{"true"}, "true (boolean)"},
		{"BooleanValue false", BooleanValue{"false"}, "false (boolean)"},
		{"NumberValue unsigned", NumberValue{"42"}, "42 (number)"},
		{"NumberValue signed", NumberValue{"-7"}, "-7 (number)"},
		{"NumberValue float", NumberValue{"3.14"}, "3.14 (number)"},
		{"StringValue quoted", StringValue{`"hello"`}, `"hello" (string)`},
		{"IdentifierValue", IdentifierValue{"fooBar"}, "fooBar (identifier)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}
