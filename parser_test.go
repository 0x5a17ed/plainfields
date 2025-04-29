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
			PositionalValueEvent{TokenIdentifier, "name"},
		}},

		{"multiple positional value", `name,123,true,false,nil,"Hello World!"`, []ParserEvent{
			PositionalValueEvent{TokenIdentifier, "name"},
			PositionalValueEvent{TokenNumber, "123"},
			PositionalValueEvent{TokenTrue, "true"},
			PositionalValueEvent{TokenFalse, "false"},
			PositionalValueEvent{TokenNil, "nil"},
			PositionalValueEvent{TokenString, `"Hello World!"`},
		}},

		{"simple field", "name=john", []ParserEvent{
			FieldStartEvent{"name"},
			ListStartEvent{},
			ValueEvent{TokenIdentifier, "john"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"field with prefix ^", "^enabled", []ParserEvent{
			FieldStartEvent{"enabled"},
			ListStartEvent{},
			ValueEvent{TokenTrue, "true"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"field with prefix !", "!disabled", []ParserEvent{
			FieldStartEvent{"disabled"},
			ListStartEvent{},
			ValueEvent{TokenFalse, "false"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"multiple fields", "name=john, age=30, active=true", []ParserEvent{
			FieldStartEvent{"name"},
			ListStartEvent{},
			ValueEvent{TokenIdentifier, "john"},
			ListEndEvent{},
			FieldEndEvent{},
			FieldStartEvent{"age"},
			ListStartEvent{},
			ValueEvent{TokenNumber, "30"},
			ListEndEvent{},
			FieldEndEvent{},
			FieldStartEvent{"active"},
			ListStartEvent{},
			ValueEvent{TokenTrue, "true"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"list values", "colors=red;blue;green", []ParserEvent{
			FieldStartEvent{"colors"},
			ListStartEvent{},
			ValueEvent{TokenIdentifier, "red"},
			ValueEvent{TokenIdentifier, "blue"},
			ValueEvent{TokenIdentifier, "green"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"map values", "settings=host:localhost;port:8080", []ParserEvent{
			FieldStartEvent{"settings"},
			MapStartEvent{},
			MapKeyEvent{TokenIdentifier, "host"},
			ValueEvent{TokenIdentifier, "localhost"},
			MapKeyEvent{TokenIdentifier, "port"},
			ValueEvent{TokenNumber, "8080"},
			MapEndEvent{},
			FieldEndEvent{},
		}},
		{"mixed value types", `data="hello";123;true;nil`, []ParserEvent{
			FieldStartEvent{"data"},
			ListStartEvent{},
			ValueEvent{TokenString, `"hello"`},
			ValueEvent{TokenNumber, "123"},
			ValueEvent{TokenTrue, "true"},
			ValueEvent{TokenNil, "nil"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"hex number", "value=0xFF", []ParserEvent{
			FieldStartEvent{"value"},
			ListStartEvent{},
			ValueEvent{TokenNumber, "0xFF"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"binary number", "flags=0b1010", []ParserEvent{
			FieldStartEvent{"flags"},
			ListStartEvent{},
			ValueEvent{TokenNumber, "0b1010"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"octal number", "perms=0o755", []ParserEvent{
			FieldStartEvent{"perms"},
			ListStartEvent{},
			ValueEvent{TokenNumber, "0o755"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"negative number", "temp=-42.5", []ParserEvent{
			FieldStartEvent{"temp"},
			ListStartEvent{},
			ValueEvent{TokenNumber, "-42.5"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"scientific notation", "value=1.23e-4", []ParserEvent{
			FieldStartEvent{"value"},
			ListStartEvent{},
			ValueEvent{TokenNumber, "1.23e-4"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"single quoted string", "msg='hello world'", []ParserEvent{
			FieldStartEvent{"msg"},
			ListStartEvent{},
			ValueEvent{TokenString, "'hello world'"},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"empty string", `empty=""`, []ParserEvent{
			FieldStartEvent{"empty"},
			ListStartEvent{},
			ValueEvent{TokenString, `""`},
			ListEndEvent{},
			FieldEndEvent{},
		}},
		{"complex example", "^enabled, name=john, settings=theme:dark;fontSize:14;autoSave:true, tags=dev;prod", []ParserEvent{
			FieldStartEvent{"enabled"},
			ListStartEvent{},
			ValueEvent{TokenTrue, "true"},
			ListEndEvent{},
			FieldEndEvent{},
			FieldStartEvent{"name"},
			ListStartEvent{},
			ValueEvent{TokenIdentifier, "john"},
			ListEndEvent{},
			FieldEndEvent{},
			FieldStartEvent{"settings"},
			MapStartEvent{},
			MapKeyEvent{TokenIdentifier, "theme"},
			ValueEvent{TokenIdentifier, "dark"},
			MapKeyEvent{TokenIdentifier, "fontSize"},
			ValueEvent{TokenNumber, "14"},
			MapKeyEvent{TokenIdentifier, "autoSave"},
			ValueEvent{TokenTrue, "true"},
			MapEndEvent{},
			FieldEndEvent{},
			FieldStartEvent{"tags"},
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
		{"missing value after assignment", "name=", "expected value, got EOF"},
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
			wantedError: `positional value "name" not allowed here`,
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
