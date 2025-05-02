package plainfields

import (
	"testing"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name    string
		builder func(b *Builder) *Builder
		options *BuilderOptions
		wanted  string
		wantErr string
	}{
		{
			name: "basic fields with defaults",
			builder: func(b *Builder) *Builder {
				return b.Enable("feature").
					Labeled("name", "john").
					Labeled("age", 30)
			},
			wanted: "^feature,name=john,age=30",
		},
		{
			name: "all field types with defaults",
			builder: func(b *Builder) *Builder {
				return b.Enable("enabled").
					Disable("debug").
					Labeled("name", "john").
					LabeledList("tags", "dev", "prod").
					LabeledDict("settings", "theme", "dark", "fontSize", 14)
			},
			wanted: "^enabled,!debug,name=john,tags=dev;prod,settings=theme:dark;fontSize:14",
		},
		{
			name: "boolean method usage",
			builder: func(b *Builder) *Builder {
				return b.Boolean("feature1", true).
					Boolean("feature2", false)
			},
			wanted: "^feature1,!feature2",
		},
		{
			name: "string escaping",
			builder: func(b *Builder) *Builder {
				return b.Labeled("simple", "abc").
					Labeled("spaces", "hello world").
					Labeled("special", "a:b;c,d").
					Labeled("empty", "").
					Labeled("keyword", "true")
			},
			wanted: `simple=abc,spaces="hello world",special="a:b;c,d",empty="",keyword="true"`,
		},
		{
			name: "empty list and pairs",
			builder: func(b *Builder) *Builder {
				return b.LabeledList("empty_list").
					LabeledDict("empty_pairs")
			},
			wanted: "empty_list=,empty_pairs=",
		},
		{
			name: "mixed value types",
			builder: func(b *Builder) *Builder {
				return b.LabeledList("mixed", "string", 123, true, nil, 45.67)
			},
			wanted: "mixed=string;123;true;nil;45.67",
		},
		{
			name: "ordered values",
			builder: func(b *Builder) *Builder {
				return b.Value("a").
					Value("b").
					Value("c").
					Labeled("d", 123).
					Labeled("flag", true)
			},
			wanted: "a,b,c,d=123,flag=true",
		},

		{
			name: "always quote strings",
			builder: func(b *Builder) *Builder {
				return b.Labeled("name", "john").
					Labeled("age", 30).
					Labeled("city", "New York").
					Labeled("country", "USA")
			},
			options: &BuilderOptions{
				AlwaysQuoteStrings: true,
			},
			wanted: `name="john",age=30,city="New York",country="USA"`,
		},
		{
			name: "spaces after field separator",
			builder: func(b *Builder) *Builder {
				return b.Enable("feature").
					Labeled("name", "john").
					Labeled("age", 30)
			},
			options: &BuilderOptions{
				SpaceAfterFieldSeparator: true,
			},
			wanted: "^feature, name=john, age=30",
		},
		{
			name: "spaces after list separator",
			builder: func(b *Builder) *Builder {
				return b.LabeledList("tags", "dev", "prod", "test")
			},
			options: &BuilderOptions{
				SpaceAfterListSeparator: true,
			},
			wanted: "tags=dev; prod; test",
		},
		{
			name: "spaces after pairs separator",
			builder: func(b *Builder) *Builder {
				return b.LabeledDict("settings", "theme", "dark", "fontSize", 14)
			},
			options: &BuilderOptions{
				SpaceAfterPairsSeparator: true,
			},
			wanted: "settings=theme: dark;fontSize: 14",
		},
		{
			name: "spaces around field assignment",
			builder: func(b *Builder) *Builder {
				return b.Labeled("name", "john").
					Labeled("age", 30)
			},
			options: &BuilderOptions{
				SpaceAroundFieldAssignment: true,
			},
			wanted: "name = john,age = 30",
		},
		{
			name: "all spacing options enabled",
			builder: func(b *Builder) *Builder {
				return b.Enable("feature").
					Labeled("name", "john").
					LabeledList("tags", "dev", "prod").
					LabeledDict("settings", "theme", "dark", "fontSize", 14)
			},
			options: &BuilderOptions{
				SpaceAfterFieldSeparator:   true,
				SpaceAfterListSeparator:    true,
				SpaceAfterPairsSeparator:   true,
				SpaceAroundFieldAssignment: true,
			},
			wanted: "^feature, name = john, tags = dev; prod, settings = theme: dark; fontSize: 14",
		},

		{
			name: "error: odd number of arguments to LabeledDict",
			builder: func(b *Builder) *Builder {
				return b.LabeledDict("settings", "key1", "value1", "key2") // Missing value
			},
			wanted:  "",
			wantErr: "odd number of arguments to LabeledDict",
		},
		{
			name: "error: ordered value after labeled field",
			builder: func(b *Builder) *Builder {
				return b.Labeled("name", "john").Value(123) // Positional value after field
			},
			wanted:  "",
			wantErr: "ordered value after labeled field",
		},
		{
			name: "error: invalid field name",
			builder: func(b *Builder) *Builder {
				return b.Labeled("invalid field", "value") // Invalid field name
			},
			wanted:  "",
			wantErr: `"invalid field": invalid field name`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var builder *Builder
			if tt.options != nil {
				builder = NewBuilder(*tt.options)
			} else {
				builder = NewBuilder()
			}
			result := tt.builder(builder).String()

			if result != tt.wanted {
				t.Errorf("expected: %q, got: %q", tt.wanted, result)
			}

			if err := builder.Err(); (err != nil) != (tt.wantErr != "") {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestNeedsQuoting(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"abc", false},        // Simple identifier
		{"", true},            // Empty string
		{"hello world", true}, // Contains space
		{"a,b", true},         // Contains comma
		{"a;b", true},         // Contains semicolon
		{"a:b", true},         // Contains colon
		{"a=b", true},         // Contains equals
		{"a\\b", true},        // Contains backslash
		{"true", true},        // Keyword
		{"false", true},       // Keyword
		{"nil", true},         // Keyword
		{"abc123", false},     // Alphanumeric
		{"abc-123", false},    // With hyphen
		{"123abc", false},     // Starts with number
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NeedsQuoting(tt.input)
			if result != tt.expected {
				t.Errorf("NeedsQuoting(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
