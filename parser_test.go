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
			ValueEvent{Value{IdentifierValue, "name"}},
			ListEndEvent{},
		}},

		{"multiple ordered value", `name,123,true,false,nil,"Hello World!"`, []ParserEvent{
			ListStartEvent{},
			ValueEvent{Value{IdentifierValue, "name"}},
			ValueEvent{Value{NumberValue, "123"}},
			ValueEvent{Value{BooleanValue, "true"}},
			ValueEvent{Value{BooleanValue, "false"}},
			ValueEvent{Value{NilValue, "nil"}},
			ValueEvent{Value{StringValue, `"Hello World!"`}},
			ListEndEvent{},
		}},

		{"ordered zero value", ",", []ParserEvent{
			ListStartEvent{},
			ValueEvent{Value{ZeroValue, ""}},
			ListEndEvent{},
		}},

		{"ordered zero value", ",,", []ParserEvent{
			ListStartEvent{},
			ValueEvent{Value{ZeroValue, ""}},
			ValueEvent{Value{ZeroValue, ""}},
			ListEndEvent{},
		}},

		{"simple labeled field", "name=john", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "name"}},
			ValueEvent{Value{IdentifierValue, "john"}},
			MapEndEvent{},
		}},

		{"ordered and labeled fields", "john, age=30", []ParserEvent{
			ListStartEvent{},
			ValueEvent{Value{IdentifierValue, "john"}},
			ListEndEvent{},
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "age"}},
			ValueEvent{Value{NumberValue, "30"}},
			MapEndEvent{},
		}},

		{"empty assignment single", "name=", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "name"}},
			ValueEvent{Value{ZeroValue, ""}},
			MapEndEvent{},
		}},

		{"empty assignment multi", "a=,b=", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "a"}},
			ValueEvent{Value{ZeroValue, ""}},
			MapKeyEvent{Value{IdentifierValue, "b"}},
			ValueEvent{Value{ZeroValue, ""}},
			MapEndEvent{},
		}},

		{"field with prefix ^", "^enabled", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "enabled"}},
			ValueEvent{Value{BooleanValue, "true"}},
			MapEndEvent{},
		}},
		{"field with prefix !", "!disabled", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "disabled"}},
			ValueEvent{Value{BooleanValue, "false"}},
			MapEndEvent{},
		}},
		{"multiple fields", "name=john, age=30, active=true", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "name"}},
			ValueEvent{Value{IdentifierValue, "john"}},
			MapKeyEvent{Value{IdentifierValue, "age"}},
			ValueEvent{Value{NumberValue, "30"}},
			MapKeyEvent{Value{IdentifierValue, "active"}},
			ValueEvent{Value{BooleanValue, "true"}},
			MapEndEvent{},
		}},
		{"list values", "colors=red;blue;green", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "colors"}},
			ListStartEvent{},
			ValueEvent{Value{IdentifierValue, "red"}},
			ValueEvent{Value{IdentifierValue, "blue"}},
			ValueEvent{Value{IdentifierValue, "green"}},
			ListEndEvent{},
			MapEndEvent{},
		}},
		{"map values", "settings=host:localhost;port:8080", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "settings"}},
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "host"}},
			ValueEvent{Value{IdentifierValue, "localhost"}},
			MapKeyEvent{Value{IdentifierValue, "port"}},
			ValueEvent{Value{NumberValue, "8080"}},
			MapEndEvent{},
			MapEndEvent{},
		}},
		{"map with boolean prefix", "features=^enabled;!disabled", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "features"}},
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "enabled"}},
			ValueEvent{Value{BooleanValue, "true"}},
			MapKeyEvent{Value{IdentifierValue, "disabled"}},
			ValueEvent{Value{BooleanValue, "false"}},
			MapEndEvent{},
			MapEndEvent{},
		}},
		{"mixed value types", `data="hello";123;true;nil`, []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "data"}},
			ListStartEvent{},
			ValueEvent{Value{StringValue, `"hello"`}},
			ValueEvent{Value{NumberValue, "123"}},
			ValueEvent{Value{BooleanValue, "true"}},
			ValueEvent{Value{NilValue, "nil"}},
			ListEndEvent{},
			MapEndEvent{},
		}},
		{"hex number", "value=0xFF", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "value"}},
			ValueEvent{Value{NumberValue, "0xFF"}},
			MapEndEvent{},
		}},
		{"binary number", "flags=0b1010", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "flags"}},
			ValueEvent{Value{NumberValue, "0b1010"}},
			MapEndEvent{},
		}},
		{"octal number", "perms=0o755", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "perms"}},
			ValueEvent{Value{NumberValue, "0o755"}},
			MapEndEvent{},
		}},
		{"negative number", "temp=-42.5", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "temp"}},
			ValueEvent{Value{NumberValue, "-42.5"}},
			MapEndEvent{},
		}},
		{"scientific notation", "value=1.23e-4", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "value"}},
			ValueEvent{Value{NumberValue, "1.23e-4"}},
			MapEndEvent{},
		}},
		{"single quoted string", "msg='hello world'", []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "msg"}},
			ValueEvent{Value{StringValue, `'hello world'`}},
			MapEndEvent{},
		}},
		{"empty string", `empty=""`, []ParserEvent{
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "empty"}},
			ValueEvent{Value{StringValue, `""`}},
			MapEndEvent{},
		}},
		{"complex example", "^enabled, name=john, settings=theme:dark;fontSize:14;autoSave:true, tags=dev;prod", []ParserEvent{
			MapStartEvent{},

			MapKeyEvent{Value{IdentifierValue, "enabled"}},
			ValueEvent{Value{BooleanValue, "true"}},

			MapKeyEvent{Value{IdentifierValue, "name"}},
			ValueEvent{Value{IdentifierValue, "john"}},

			MapKeyEvent{Value{IdentifierValue, "settings"}},
			MapStartEvent{},
			MapKeyEvent{Value{IdentifierValue, "theme"}},
			ValueEvent{Value{IdentifierValue, "dark"}},
			MapKeyEvent{Value{IdentifierValue, "fontSize"}},
			ValueEvent{Value{NumberValue, "14"}},
			MapKeyEvent{Value{IdentifierValue, "autoSave"}},
			ValueEvent{Value{BooleanValue, "true"}},
			MapEndEvent{},

			MapKeyEvent{Value{IdentifierValue, "tags"}},
			ListStartEvent{},
			ValueEvent{Value{IdentifierValue, "dev"}},
			ValueEvent{Value{IdentifierValue, "prod"}},
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
		ValueEvent{Value{Type: StringValue, Value: `"test"`}},
		ListStartEvent{},
		ListEndEvent{},
		MapStartEvent{},
		MapEndEvent{},
		MapKeyEvent{Value{Type: IdentifierValue, Value: "key"}},
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
