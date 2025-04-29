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
		// We expect success with a specific value
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

func TestValueEventConversions(t *testing.T) {
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
		{input: `-42`, wantInt: p(int64(-42)), wantFloat: p(float64(-42))},
		{input: `23`, wantInt: p(int64(23)), wantUint: p(uint64(23)), wantFloat: p(float64(23))},
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
