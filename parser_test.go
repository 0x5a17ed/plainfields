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

		{"single positional value", "name", []ParserEvent{
			OrderedFieldStartEvent{0},
			ValueEvent{TokenIdentifier, "name"},
			FieldEndEvent{},
		}},

		{"multiple positional value", `name,123,true,false,nil,"Hello World!"`, []ParserEvent{
			OrderedFieldStartEvent{0},
			ValueEvent{TokenIdentifier, "name"},
			FieldEndEvent{},
			OrderedFieldStartEvent{1},
			ValueEvent{TokenNumber, "123"},
			FieldEndEvent{},
			OrderedFieldStartEvent{2},
			ValueEvent{TokenTrue, "true"},
			FieldEndEvent{},
			OrderedFieldStartEvent{3},
			ValueEvent{TokenFalse, "false"},
			FieldEndEvent{},
			OrderedFieldStartEvent{4},
			ValueEvent{TokenNil, "nil"},
			FieldEndEvent{},
			OrderedFieldStartEvent{5},
			ValueEvent{TokenString, `"Hello World!"`},
			FieldEndEvent{},
		}},

		{"simple field", "name=john", []ParserEvent{
			LabeledFieldStartEvent{"name"},
			ValueEvent{TokenIdentifier, "john"},
			FieldEndEvent{},
		}},

		{"empty assignment single", "name=", []ParserEvent{
			LabeledFieldStartEvent{"name"},
			FieldEndEvent{},
		}},

		{"empty assignment multi", "a=,b=", []ParserEvent{
			LabeledFieldStartEvent{"a"},
			FieldEndEvent{},
			LabeledFieldStartEvent{"b"},
			FieldEndEvent{},
		}},

		{"field with prefix ^", "^enabled", []ParserEvent{
			LabeledFieldStartEvent{"enabled"},
			ValueEvent{TokenTrue, "true"},
			FieldEndEvent{},
		}},
		{"field with prefix !", "!disabled", []ParserEvent{
			LabeledFieldStartEvent{"disabled"},
			ValueEvent{TokenFalse, "false"},
			FieldEndEvent{},
		}},
		{"multiple fields", "name=john, age=30, active=true", []ParserEvent{
			LabeledFieldStartEvent{"name"},
			ValueEvent{TokenIdentifier, "john"},
			FieldEndEvent{},
			LabeledFieldStartEvent{"age"},
			ValueEvent{TokenNumber, "30"},
			FieldEndEvent{},
			LabeledFieldStartEvent{"active"},
			ValueEvent{TokenTrue, "true"},
			FieldEndEvent{},
		}},
		{"list values", "colors=red;blue;green", []ParserEvent{
			LabeledFieldStartEvent{"colors"},
			ListStartEvent{},
			ValueEvent{TokenIdentifier, "red"},
			ValueEvent{TokenIdentifier, "blue"},
			ValueEvent{TokenIdentifier, "green"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"map values", "settings=host:localhost;port:8080", []ParserEvent{
			LabeledFieldStartEvent{"settings"},
			MapStartEvent{},
			MapKeyEvent{TokenIdentifier, "host"},
			ValueEvent{TokenIdentifier, "localhost"},
			MapKeyEvent{TokenIdentifier, "port"},
			ValueEvent{TokenNumber, "8080"},
			MapEndEvent{},
			FieldEndEvent{},
		}},
		{"mixed value types", `data="hello";123;true;nil`, []ParserEvent{
			LabeledFieldStartEvent{"data"},
			ListStartEvent{},
			ValueEvent{TokenString, `"hello"`},
			ValueEvent{TokenNumber, "123"},
			ValueEvent{TokenTrue, "true"},
			ValueEvent{TokenNil, "nil"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"hex number", "value=0xFF", []ParserEvent{
			LabeledFieldStartEvent{"value"},
			ValueEvent{TokenNumber, "0xFF"},
			FieldEndEvent{},
		}},
		{"binary number", "flags=0b1010", []ParserEvent{
			LabeledFieldStartEvent{"flags"},
			ValueEvent{TokenNumber, "0b1010"},
			FieldEndEvent{},
		}},
		{"octal number", "perms=0o755", []ParserEvent{
			LabeledFieldStartEvent{"perms"},
			ValueEvent{TokenNumber, "0o755"},
			FieldEndEvent{},
		}},
		{"negative number", "temp=-42.5", []ParserEvent{
			LabeledFieldStartEvent{"temp"},
			ValueEvent{TokenNumber, "-42.5"},
			FieldEndEvent{},
		}},
		{"scientific notation", "value=1.23e-4", []ParserEvent{
			LabeledFieldStartEvent{"value"},
			ValueEvent{TokenNumber, "1.23e-4"},
			FieldEndEvent{},
		}},
		{"single quoted string", "msg='hello world'", []ParserEvent{
			LabeledFieldStartEvent{"msg"},
			ValueEvent{TokenString, "'hello world'"},
			FieldEndEvent{},
		}},
		{"empty string", `empty=""`, []ParserEvent{
			LabeledFieldStartEvent{"empty"},
			ValueEvent{TokenString, `""`},
			FieldEndEvent{},
		}},
		{"complex example", "^enabled, name=john, settings=theme:dark;fontSize:14;autoSave:true, tags=dev;prod", []ParserEvent{
			LabeledFieldStartEvent{"enabled"},
			ValueEvent{TokenTrue, "true"},
			FieldEndEvent{},

			LabeledFieldStartEvent{"name"},
			ValueEvent{TokenIdentifier, "john"},
			FieldEndEvent{},

			LabeledFieldStartEvent{"settings"},
			MapStartEvent{},
			MapKeyEvent{TokenIdentifier, "theme"},
			ValueEvent{TokenIdentifier, "dark"},
			MapKeyEvent{TokenIdentifier, "fontSize"},
			ValueEvent{TokenNumber, "14"},
			MapKeyEvent{TokenIdentifier, "autoSave"},
			ValueEvent{TokenTrue, "true"},
			MapEndEvent{},
			FieldEndEvent{},

			LabeledFieldStartEvent{"tags"},
			ListStartEvent{},
			ValueEvent{TokenIdentifier, "dev"},
			ValueEvent{TokenIdentifier, "prod"},
			ListEndEvent{},
			FieldEndEvent{},
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
				t.Errorf("ParseTokens() = %v, want %v", got, tt.expected)
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
		{"double field seperator", "a=1,,", "expected field prefix, identifier, or value, got FieldSeparator"},
		{"invalid value", "name==", "expected value, got Assign"},
		{"incomplete map", "settings=key:", "expected value, got EOF"},
		{"map missing key after list separator", "settings=key:value;", "expected value, got EOF"},
		{"mixing map and list semantics", "settings=key:value;value", "expected PairSeparator, got EOF"},
		{"missing value after list separator", "a=1;", "expected value, got EOF"},
		{"invalid map key", "settings=:value", "expected value, got PairSeparator"},
		{"positional value after field", "name=john,123", "positional value not allowed here"},
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
			name: "disable positional values",
			options: ParseOptions{
				AllowPositional: false,
			},
			input:       "name,omitempty",
			wantedError: `positional value not allowed here`,
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
		OrderedFieldStartEvent{Index: 42},
		LabeledFieldStartEvent{Name: "foo"},
		FieldEndEvent{},
		ValueEvent{Type: TokenString, Value: "test"},
		ListStartEvent{},
		ListEndEvent{},
		MapStartEvent{},
		MapEndEvent{},
		MapKeyEvent{Type: TokenIdentifier, Value: "key"},
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
