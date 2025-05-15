package kaval

import (
	"reflect"
	"testing"
)

// collectEvents is a helper function to collect events from the parser.
func collectEvents(input string, options ...ParseOptions) ([]ParserEvent, *ErrorEvent) {
	var events []ParserEvent

	for event := range Parse(input, options...) {
		if err, isError := event.(ErrorEvent); isError {
			return nil, &err
		} else {
			events = append(events, event)
		}
	}

	return events, nil
}

func newValue(t ValueType, v string) Value {
	switch t {
	case IdentifierValueType:
		return IdentifierValue{raw: v}
	case NumberValueType:
		return NumberValue{raw: v}
	case StringValueType:
		return StringValue{raw: v}
	case BooleanValueType:
		return BooleanValue{raw: v}
	case NilValueType:
		return NilValue{}
	default:
		return nil
	}
}

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ParserEvent
	}{
		{"empty input", "", nil},

		{"single ordered value", "name", []ParserEvent{
			ListStartEvent{},
			ValueEvent{newValue(IdentifierValueType, "name")},
			ListEndEvent{},
		}},

		{"multiple ordered value", `name,123,true,false,nil,"Hello World!"`, []ParserEvent{
			ListStartEvent{},
			ValueEvent{newValue(IdentifierValueType, "name")},
			ValueEvent{newValue(NumberValueType, "123")},
			ValueEvent{newValue(BooleanValueType, "true")},
			ValueEvent{newValue(BooleanValueType, "false")},
			ValueEvent{newValue(NilValueType, "nil")},
			ValueEvent{newValue(StringValueType, `"Hello World!"`)},
			ListEndEvent{},
		}},

		{"ordered implicit nil field", ",", []ParserEvent{
			ListStartEvent{},
			ValueEvent{newValue(NilValueType, "")},
			ListEndEvent{},
		}},

		{"ordered implicit multiple nil values", ",,", []ParserEvent{
			ListStartEvent{},
			ValueEvent{newValue(NilValueType, "")},
			ValueEvent{newValue(NilValueType, "")},
			ListEndEvent{},
		}},

		{"labeled implicit nil single field", "name=", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "name")},
			ValueEvent{newValue(NilValueType, "")},
			MapEndEvent{},
		}},

		{"labeled implicit nil multiple fields", "a=,b=", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "a")},
			ValueEvent{newValue(NilValueType, "")},
			MapKeyEvent{newValue(IdentifierValueType, "b")},
			ValueEvent{newValue(NilValueType, "")},
			MapEndEvent{},
		}},

		{"labeled field identifier", "name=john", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "name")},
			ValueEvent{newValue(IdentifierValueType, "john")},
			MapEndEvent{},
		}},

		{"ordered and labeled fields", "john, age=30", []ParserEvent{
			ListStartEvent{},
			ValueEvent{newValue(IdentifierValueType, "john")},
			ListEndEvent{},
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "age")},
			ValueEvent{newValue(NumberValueType, "30")},
			MapEndEvent{},
		}},

		{"field with prefix ^", "^enabled", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "enabled")},
			ValueEvent{newValue(BooleanValueType, "true")},
			MapEndEvent{},
		}},
		{"field with prefix !", "!disabled", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "disabled")},
			ValueEvent{newValue(BooleanValueType, "false")},
			MapEndEvent{},
		}},
		{"multiple fields", "name=john, age=30, active=true", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "name")},
			ValueEvent{newValue(IdentifierValueType, "john")},
			MapKeyEvent{newValue(IdentifierValueType, "age")},
			ValueEvent{newValue(NumberValueType, "30")},
			MapKeyEvent{newValue(IdentifierValueType, "active")},
			ValueEvent{newValue(BooleanValueType, "true")},
			MapEndEvent{},
		}},

		{"ordered list value", "red;blue;green", []ParserEvent{
			ListStartEvent{},
			ListStartEvent{},
			ValueEvent{newValue(IdentifierValueType, "red")},
			ValueEvent{newValue(IdentifierValueType, "blue")},
			ValueEvent{newValue(IdentifierValueType, "green")},
			ListEndEvent{},
			ListEndEvent{},
		}},

		{"labeled list values", "colors=red;blue;green", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "colors")},
			ListStartEvent{},
			ValueEvent{newValue(IdentifierValueType, "red")},
			ValueEvent{newValue(IdentifierValueType, "blue")},
			ValueEvent{newValue(IdentifierValueType, "green")},
			ListEndEvent{},
			MapEndEvent{},
		}},

		{"ordered map values", "host:localhost;port:8080", []ParserEvent{
			ListStartEvent{},
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "host")},
			ValueEvent{newValue(IdentifierValueType, "localhost")},
			MapKeyEvent{newValue(IdentifierValueType, "port")},
			ValueEvent{newValue(NumberValueType, "8080")},
			MapEndEvent{},
			ListEndEvent{},
		}},

		{"labeled map values", "settings=host:localhost;port:8080", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "settings")},
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "host")},
			ValueEvent{newValue(IdentifierValueType, "localhost")},
			MapKeyEvent{newValue(IdentifierValueType, "port")},
			ValueEvent{newValue(NumberValueType, "8080")},
			MapEndEvent{},
			MapEndEvent{},
		}},

		{"map with boolean prefix", "features=^enabled;!disabled", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "features")},
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "enabled")},
			ValueEvent{newValue(BooleanValueType, "true")},
			MapKeyEvent{newValue(IdentifierValueType, "disabled")},
			ValueEvent{newValue(BooleanValueType, "false")},
			MapEndEvent{},
			MapEndEvent{},
		}},
		{"mixed value types", `data="hello";123;true;nil`, []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "data")},
			ListStartEvent{},
			ValueEvent{newValue(StringValueType, `"hello"`)},
			ValueEvent{newValue(NumberValueType, "123")},
			ValueEvent{newValue(BooleanValueType, "true")},
			ValueEvent{newValue(NilValueType, "nil")},
			ListEndEvent{},
			MapEndEvent{},
		}},
		{"hex number", "value=0xFF", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "value")},
			ValueEvent{newValue(NumberValueType, "0xFF")},
			MapEndEvent{},
		}},
		{"binary number", "flags=0b1010", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "flags")},
			ValueEvent{newValue(NumberValueType, "0b1010")},
			MapEndEvent{},
		}},
		{"octal number", "perms=0o755", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "perms")},
			ValueEvent{newValue(NumberValueType, "0o755")},
			MapEndEvent{},
		}},
		{"negative number", "temp=-42.5", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "temp")},
			ValueEvent{newValue(NumberValueType, "-42.5")},
			MapEndEvent{},
		}},
		{"scientific notation", "value=1.23e-4", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "value")},
			ValueEvent{newValue(NumberValueType, "1.23e-4")},
			MapEndEvent{},
		}},
		{"single quoted string", "msg='hello world'", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "msg")},
			ValueEvent{newValue(StringValueType, `"hello world"`)},
			MapEndEvent{},
		}},
		{"empty string", `empty=""`, []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "empty")},
			ValueEvent{newValue(StringValueType, `""`)},
			MapEndEvent{},
		}},
		{"complex example", "^enabled, name=john, settings=theme:dark;fontSize:14;autoSave:true, tags=dev;prod", []ParserEvent{
			MapStartEvent{},

			MapKeyEvent{newValue(IdentifierValueType, "enabled")},
			ValueEvent{newValue(BooleanValueType, "true")},

			MapKeyEvent{newValue(IdentifierValueType, "name")},
			ValueEvent{newValue(IdentifierValueType, "john")},

			MapKeyEvent{newValue(IdentifierValueType, "settings")},
			MapStartEvent{},
			MapKeyEvent{newValue(IdentifierValueType, "theme")},
			ValueEvent{newValue(IdentifierValueType, "dark")},
			MapKeyEvent{newValue(IdentifierValueType, "fontSize")},
			ValueEvent{newValue(NumberValueType, "14")},
			MapKeyEvent{newValue(IdentifierValueType, "autoSave")},
			ValueEvent{newValue(BooleanValueType, "true")},
			MapEndEvent{},

			MapKeyEvent{newValue(IdentifierValueType, "tags")},
			ListStartEvent{},
			ValueEvent{newValue(IdentifierValueType, "dev")},
			ValueEvent{newValue(IdentifierValueType, "prod")},
			ListEndEvent{},

			MapEndEvent{},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collectEvents(tt.input)
			if err != nil {
				t.Errorf("ParseTokens() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseTokens() = %#v, want %#v", got, tt.expected)
			}
		})
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError string
	}{
		{"missing identifier after prefix", "^", "expected Identifier, got EOF"},
		{"double field seperator", "a=1,,", "ordered value not allowed here"},
		{"invalid value", "name==", "expected value, got Assign"},
		{"incomplete map", "settings=key:", "expected value, got EOF"},
		{"map missing key after list separator", "settings=key:value;", "expected value, got EOF"},
		{"mixing map and list semantics", "settings=key:value;value", "expected PairSeparator, got EOF"},
		{"missing value after list separator", "a=1;", "expected value, got EOF"},
		{"invalid map key", "settings=:value", "expected value, got PairSeparator"},
		{"ordered field after labeled field", "name=john,123", "ordered value not allowed here"},
		{"invalid boolean prefix", "^=true", "expected Identifier, got Assign"},
		{"invalid boolean prefix with space", "^ =true", "expected Identifier, got Assign"},
		{"invalid boolean prefix with extra token", "^enabled,=true", "expected field prefix, identifier, or value, got Assign"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := collectEvents(tt.input); err == nil {
				t.Errorf("Expected error, got none")
			} else if err.Msg != tt.expectError {
				t.Errorf("ParseTokens() error = %q, want %q", err.Msg, tt.expectError)
			}
		})
	}
}

func TestParseOptions(t *testing.T) {
	tt := []struct {
		name         string
		options      ParseOptions
		input        string
		wantedEvents []ParserEvent
		wantedError  string
	}{
		{
			name: "disable ordered values",
			options: ParseOptions{
				AllowOrdered: false,
			},
			input:       "name,omitempty",
			wantedError: `ordered value not allowed here`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := collectEvents(tc.input, tc.options)

			if err != nil && tc.wantedError == "" {
				t.Errorf("ParseWithOptions() error = %v", err)
			} else if err == nil && tc.wantedError != "" {
				t.Errorf("ParseWithOptions() expected error = %q, got none", tc.wantedError)
			} else if err != nil && tc.wantedError != "" && err.Msg != tc.wantedError {
				t.Errorf("ParseWithOptions() error = %q, want %q", err.Msg, tc.wantedError)
			} else if err == nil && tc.wantedError == "" {
				if !reflect.DeepEqual(got, tc.wantedEvents) {
					t.Errorf("ParseWithOptions() = %v, want %v", got, tc.wantedEvents)
				}
			}

		})
	}
}

// TestParserEventInterface is a silly test that simply calls isParserEvent() on each
// event type to improve test coverage and doesn't test any functionality.
func TestParserEventInterface(t *testing.T) {
	// Create instances of each event type.
	events := []ParserEvent{
		ValueEvent{newValue(StringValueType, `"test"`)},
		ListStartEvent{},
		ListEndEvent{},
		MapStartEvent{},
		MapEndEvent{},
		MapKeyEvent{newValue(IdentifierValueType, "key")},
		ErrorEvent{Msg: "error", Pos: Position{Offset: 0, Column: 1}},
	}

	for _, event := range events {
		// This doesn't do anything meaningful at runtime,
		// but it ensures the method is called for test coverage.
		event.isParserEvent()
	}

	t.Run("error interface", func(t *testing.T) {
		var err error = ErrorEvent{Msg: "test error"}
		if err.Error() != "Error at Col 0 (Offset 0): test error" {
			t.Errorf("Expected error message to be 'test error', got '%s'", err.Error())
		}
	})
}
