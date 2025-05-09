package plainfields

import (
	"fmt"
	"testing"
)

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
func testConversion[T comparable](t *testing.T, name string, want *T, fn func() (T, error)) {
	if got, err := fn(); want != nil {
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

func TestValueEvent_Conversions(t *testing.T) {
	tests := []struct {
		input string

		wantString *string
		wantInt    *int64
		wantUint   *uint64
		wantFloat  *float64
		wantBool   *bool
		wantNil    bool
	}{
		{input: `field`, wantString: p("field")},
		{input: `"hello"`, wantString: p("hello")},
		{input: `-42`, wantString: p("-42"), wantInt: p(int64(-42)), wantFloat: p(float64(-42))},
		{input: `23`, wantString: p("23"), wantInt: p(int64(23)), wantUint: p(uint64(23)), wantFloat: p(float64(23))},
		{input: `3.14`, wantString: p("3.14"), wantFloat: p(float64(3.14))},

		{input: "0x23", wantString: p("0x23"), wantInt: p(int64(0x23)), wantUint: p(uint64(0x23)), wantFloat: p(float64(0x23))},
		{input: "0x23.1", wantString: p("0x23.1"), wantFloat: p(float64(35.0625))},
		{input: "0x1p1", wantString: p("0x1p1"), wantFloat: p(float64(2))},
		{input: "0x1.8p1", wantString: p("0x1.8p1"), wantFloat: p(float64(3.0))},
		{input: "0x1.8p-1", wantString: p("0x1.8p-1"), wantFloat: p(float64(0.75))},
		{input: "0x1.8p+1", wantString: p("0x1.8p+1"), wantFloat: p(float64(3.0))},

		{input: "0o644", wantString: p("0o644"), wantInt: p(int64(420)), wantUint: p(uint64(420)), wantFloat: p(float64(420))},
		{input: "0b01100101", wantString: p("0b01100101"), wantInt: p(int64(101)), wantUint: p(uint64(101)), wantFloat: p(float64(101))},

		{input: `true`, wantBool: p(true)},
		{input: `false`, wantBool: p(false)},
		{input: `nil`, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ev, err := collectEvent[ValueEvent](tt.input)
			if err != nil {
				t.Fatalf("collectEvent() error = %v", err)
			}

			testConversion(t, "ToString", tt.wantString, ev.ToString)
			testConversion(t, "ToInt", tt.wantInt, ev.ToInt)
			testConversion(t, "ToUint", tt.wantUint, ev.ToUint)
			testConversion(t, "ToFloat", tt.wantFloat, ev.ToFloat)
			testConversion(t, "ToBool", tt.wantBool, ev.ToBool)

			if got := ev.IsNil(); got != tt.wantNil {
				t.Errorf("IsNil() = %t, want %t", got, tt.wantNil)
			}
		})
	}
}

func TestValueEvent_FloatError(t *testing.T) {
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
			ev := Value{Type: NumberValueType, Value: tt.input}

			_, err := ev.ToFloat()
			if err == nil {
				t.Fatalf("ToFloat() error = %v", err)
			}
		})
	}
}
