package plainfields

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

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ParserEvent
	}{
		{"empty input", "", nil},

		{"single ordered value", "name", []ParserEvent{
			ListStartEvent{},
			ValueEvent{Value{IdentifierValueType, "name"}},
			ListEndEvent{},
		}},

		{"multiple ordered value", `name,123,true,false,nil,"Hello World!"`, []ParserEvent{
			ListStartEvent{},
			ValueEvent{Value{IdentifierValueType, "name"}},
			ValueEvent{Value{NumberValueType, "123"}},
			ValueEvent{Value{BooleanValueType, "true"}},
			ValueEvent{Value{BooleanValueType, "false"}},
			ValueEvent{Value{NilValueType, "nil"}},
			ValueEvent{Value{StringValueType, `"Hello World!"`}},
			ListEndEvent{},
		}},

		{"ordered zero value", ",", []ParserEvent{
			ListStartEvent{},
			ValueEvent{Value{ZeroValueType, ""}},
			ListEndEvent{},
		}},

		{"ordered zero value", ",,", []ParserEvent{
			ListStartEvent{},
			ValueEvent{Value{ZeroValueType, ""}},
			ValueEvent{Value{ZeroValueType, ""}},
			ListEndEvent{},
		}},

		{"labeled field identifier", "name=john", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "name"}},
			ValueEvent{Value{IdentifierValueType, "john"}},
			MapEndEvent{},
		}},

		{"ordered and labeled fields", "john, age=30", []ParserEvent{
			ListStartEvent{},
			ValueEvent{Value{IdentifierValueType, "john"}},
			ListEndEvent{},
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "age"}},
			ValueEvent{Value{NumberValueType, "30"}},
			MapEndEvent{},
		}},

		{"empty assignment single", "name=", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "name"}},
			ValueEvent{Value{ZeroValueType, ""}},
			MapEndEvent{},
		}},

		{"empty assignment multi", "a=,b=", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "a"}},
			ValueEvent{Value{ZeroValueType, ""}},
			MapKeyEvent{Value{IdentifierValueType, "b"}},
			ValueEvent{Value{ZeroValueType, ""}},
			MapEndEvent{},
		}},

		{"field with prefix ^", "^enabled", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "enabled"}},
			ValueEvent{Value{BooleanValueType, "true"}},
			MapEndEvent{},
		}},
		{"field with prefix !", "!disabled", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "disabled"}},
			ValueEvent{Value{BooleanValueType, "false"}},
			MapEndEvent{},
		}},
		{"multiple fields", "name=john, age=30, active=true", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "name"}},
			ValueEvent{Value{IdentifierValueType, "john"}},
			MapKeyEvent{Value{IdentifierValueType, "age"}},
			ValueEvent{Value{NumberValueType, "30"}},
			MapKeyEvent{Value{IdentifierValueType, "active"}},
			ValueEvent{Value{BooleanValueType, "true"}},
			MapEndEvent{},
		}},

		{"labeled list values", "colors=red;blue;green", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "colors"}},
			ListStartEvent{},
			ValueEvent{Value{IdentifierValueType, "red"}},
			ValueEvent{Value{IdentifierValueType, "blue"}},
			ValueEvent{Value{IdentifierValueType, "green"}},
			ListEndEvent{},
			MapEndEvent{},
		}},

		{"labeled map values", "settings=host:localhost;port:8080", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "settings"}},
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "host"}},
			ValueEvent{Value{IdentifierValueType, "localhost"}},
			MapKeyEvent{Value{IdentifierValueType, "port"}},
			ValueEvent{Value{NumberValueType, "8080"}},
			MapEndEvent{},
			MapEndEvent{},
		}},

		{"map with boolean prefix", "features=^enabled;!disabled", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "features"}},
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "enabled"}},
			ValueEvent{Value{BooleanValueType, "true"}},
			MapKeyEvent{Value{IdentifierValueType, "disabled"}},
			ValueEvent{Value{BooleanValueType, "false"}},
			MapEndEvent{},
			MapEndEvent{},
		}},
		{"mixed value types", `data="hello";123;true;nil`, []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "data"}},
			ListStartEvent{},
			ValueEvent{Value{StringValueType, `"hello"`}},
			ValueEvent{Value{NumberValueType, "123"}},
			ValueEvent{Value{BooleanValueType, "true"}},
			ValueEvent{Value{NilValueType, "nil"}},
			ListEndEvent{},
			MapEndEvent{},
		}},
		{"hex number", "value=0xFF", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "value"}},
			ValueEvent{Value{NumberValueType, "0xFF"}},
			MapEndEvent{},
		}},
		{"binary number", "flags=0b1010", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "flags"}},
			ValueEvent{Value{NumberValueType, "0b1010"}},
			MapEndEvent{},
		}},
		{"octal number", "perms=0o755", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "perms"}},
			ValueEvent{Value{NumberValueType, "0o755"}},
			MapEndEvent{},
		}},
		{"negative number", "temp=-42.5", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "temp"}},
			ValueEvent{Value{NumberValueType, "-42.5"}},
			MapEndEvent{},
		}},
		{"scientific notation", "value=1.23e-4", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "value"}},
			ValueEvent{Value{NumberValueType, "1.23e-4"}},
			MapEndEvent{},
		}},
		{"single quoted string", "msg='hello world'", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "msg"}},
			ValueEvent{Value{StringValueType, `'hello world'`}},
			MapEndEvent{},
		}},
		{"empty string", `empty=""`, []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "empty"}},
			ValueEvent{Value{StringValueType, `""`}},
			MapEndEvent{},
		}},
		{"complex example", "^enabled, name=john, settings=theme:dark;fontSize:14;autoSave:true, tags=dev;prod", []ParserEvent{
			MapStartEvent{},

			MapKeyEvent{Value{IdentifierValueType, "enabled"}},
			ValueEvent{Value{BooleanValueType, "true"}},

			MapKeyEvent{Value{IdentifierValueType, "name"}},
			ValueEvent{Value{IdentifierValueType, "john"}},

			MapKeyEvent{Value{IdentifierValueType, "settings"}},
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValueType, "theme"}},
			ValueEvent{Value{IdentifierValueType, "dark"}},
			MapKeyEvent{Value{IdentifierValueType, "fontSize"}},
			ValueEvent{Value{NumberValueType, "14"}},
			MapKeyEvent{Value{IdentifierValueType, "autoSave"}},
			ValueEvent{Value{BooleanValueType, "true"}},
			MapEndEvent{},

			MapKeyEvent{Value{IdentifierValueType, "tags"}},
			ListStartEvent{},
			ValueEvent{Value{IdentifierValueType, "dev"}},
			ValueEvent{Value{IdentifierValueType, "prod"}},
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
		ValueEvent{Value{Type: StringValueType, Value: `"test"`}},
		ListStartEvent{},
		ListEndEvent{},
		MapStartEvent{},
		MapEndEvent{},
		MapKeyEvent{Value{Type: IdentifierValueType, Value: "key"}},
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
