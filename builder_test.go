package plainfields

import (
	"testing"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name    string
		builder func(b *Builder) *Builder
		options BuilderOptions
		wanted  string
	}{
		{
			name: "basic fields with defaults",
			builder: func(b *Builder) *Builder {
				return b.Enable("feature").
					Field("name", "john").
					Field("age", 30)
			},
			options: BuilderDefaults(),
			wanted:  "^feature,name=john,age=30",
		},
		{
			name: "all field types with defaults",
			builder: func(b *Builder) *Builder {
				return b.Enable("enabled").
					Disable("debug").
					Field("name", "john").
					List("tags", "dev", "prod").
					Pairs("settings", "theme", "dark", "fontSize", 14)
			},
			options: BuilderDefaults(),
			wanted:  "^enabled,!debug,name=john,tags=dev;prod,settings=theme:dark;fontSize:14",
		},
		{
			name: "spaces after field separator",
			builder: func(b *Builder) *Builder {
				return b.Enable("feature").
					Field("name", "john").
					Field("age", 30)
			},
			options: BuilderOptions{
				SpaceAfterFieldSeparator: true,
			},
			wanted: "^feature, name=john, age=30",
		},
		{
			name: "spaces after list separator",
			builder: func(b *Builder) *Builder {
				return b.List("tags", "dev", "prod", "test")
			},
			options: BuilderOptions{
				SpaceAfterListSeparator: true,
			},
			wanted: "tags=dev; prod; test",
		},
		{
			name: "spaces after pairs separator",
			builder: func(b *Builder) *Builder {
				return b.Pairs("settings", "theme", "dark", "fontSize", 14)
			},
			options: BuilderOptions{
				SpaceAfterPairsSeparator: true,
			},
			wanted: "settings=theme: dark;fontSize: 14",
		},
		{
			name: "spaces around field assignment",
			builder: func(b *Builder) *Builder {
				return b.Field("name", "john").
					Field("age", 30)
			},
			options: BuilderOptions{
				SpaceAroundFieldAssignment: true,
			},
			wanted: "name = john,age = 30",
		},
		{
			name: "all spacing options enabled",
			builder: func(b *Builder) *Builder {
				return b.Enable("feature").
					Field("name", "john").
					List("tags", "dev", "prod").
					Pairs("settings", "theme", "dark", "fontSize", 14)
			},
			options: BuilderOptions{
				SpaceAfterFieldSeparator:   true,
				SpaceAfterListSeparator:    true,
				SpaceAfterPairsSeparator:   true,
				SpaceAroundFieldAssignment: true,
			},
			wanted: "^feature, name = john, tags = dev; prod, settings = theme: dark; fontSize: 14",
		},
		{
			name: "boolean method usage",
			builder: func(b *Builder) *Builder {
				return b.Boolean("feature1", true).
					Boolean("feature2", false)
			},
			options: BuilderDefaults(),
			wanted:  "^feature1,!feature2",
		},
		{
			name: "string escaping",
			builder: func(b *Builder) *Builder {
				return b.Field("simple", "abc").
					Field("spaces", "hello world").
					Field("special", "a:b;c,d").
					Field("empty", "").
					Field("keyword", "true")
			},
			options: BuilderDefaults(),
			wanted:  `simple=abc,spaces="hello world",special="a:b;c,d",empty="",keyword="true"`,
		},
		{
			name: "empty list and pairs",
			builder: func(b *Builder) *Builder {
				return b.List("empty_list").
					Pairs("empty_pairs")
			},
			options: BuilderDefaults(),
			wanted:  "empty_list=,empty_pairs=",
		},
		{
			name: "mixed value types",
			builder: func(b *Builder) *Builder {
				return b.List("mixed", "string", 123, true, nil, 45.67)
			},
			options: BuilderDefaults(),
			wanted:  "mixed=string;123;true;nil;45.67",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilderWithOptions(tt.options)
			result := tt.builder(builder).String()

			if result != tt.wanted {
				t.Errorf("expected: %q, got: %q", tt.wanted, result)
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

// Test that the builder panics with odd number of arguments to Pairs
func TestPairsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic with odd number of pair arguments, but no panic occurred")
		}
	}()

	NewBuilder().Pairs("settings", "key1", "value1", "key2") // Missing value
}
