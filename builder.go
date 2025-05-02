package plainfields

import (
	"fmt"
	"strings"
)

// NeedsQuoting returns true if the given string needs quotes.
func NeedsQuoting(s string) bool {
	return s == "" ||
		strings.ContainsAny(s, ` ,;:=\`) ||
		s == "true" || s == "false" || s == "nil"
}

// formatValue returns a string representation suitable for plainfields.
func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		if NeedsQuoting(val) {
			return fmt.Sprintf("%q", val)
		}
		return val
	case nil:
		return "nil"
	default:
		return fmt.Sprint(v)
	}
}

// BuilderOptions controls Builder formatting behavior.
type BuilderOptions struct {
	// SpaceAfterFieldSeparator adds a space after the field separator `,`.
	SpaceAfterFieldSeparator bool

	// SpaceAfterListSeparator adds a space after the list separator `;`.
	SpaceAfterListSeparator bool

	// SpaceAfterPairsSeparator adds a space after the map key-value separator `:`.
	SpaceAfterPairsSeparator bool

	// SpaceAroundFieldAssignment adds a space around the field assignment `=`.
	SpaceAroundFieldAssignment bool
}

// BuilderDefaults returns the default formatting options
func BuilderDefaults() BuilderOptions {
	return BuilderOptions{
		SpaceAfterFieldSeparator:   false,
		SpaceAfterPairsSeparator:   false,
		SpaceAroundFieldAssignment: false,
	}
}

// Builder constructs plainfields format strings.
type Builder struct {
	options    BuilderOptions
	fields     []string
	hasLabeled bool // Track if we've seen any labeled fields.
}

// Ordered adds an ordered value to the builder.
func (b *Builder) Ordered(value any) *Builder {
	if b.hasLabeled {
		panic("ordered values must come before labeled fields")
	}

	b.fields = append(b.fields, formatValue(value))
	return b
}

// Enable adds a boolean field with ^ prefix.
func (b *Builder) Enable(name string) *Builder {
	b.hasLabeled = true
	b.fields = append(b.fields, "^"+name)
	return b
}

// Disable adds a boolean field with ! prefix.
func (b *Builder) Disable(name string) *Builder {
	b.hasLabeled = true
	b.fields = append(b.fields, "!"+name)
	return b
}

// Boolean adds a boolean field with ^ or ! prefix.
func (b *Builder) Boolean(name string, value bool) *Builder {
	if value {
		return b.Enable(name)
	}
	return b.Disable(name)
}

// Labeled adds a name=value field
func (b *Builder) Labeled(name string, value any) *Builder {
	b.hasLabeled = true

	equals := "="
	if b.options.SpaceAroundFieldAssignment {
		equals = " = "
	}

	b.fields = append(b.fields, fmt.Sprintf("%s%s%v", name, equals, formatValue(value)))
	return b
}

// List adds a name=value1;value2;... field
func (b *Builder) List(name string, values ...any) *Builder {
	b.hasLabeled = true

	equals := "="
	if b.options.SpaceAroundFieldAssignment {
		equals = " = "
	}

	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = formatValue(v)
	}

	separator := ";"
	if b.options.SpaceAfterListSeparator {
		separator = "; "
	}

	b.fields = append(b.fields, fmt.Sprintf("%s%s%s",
		name, equals, strings.Join(parts, separator)))
	return b
}

// Pairs adds a name=key1:value1;key2:value2 field
func (b *Builder) Pairs(name string, pairs ...any) *Builder {
	b.hasLabeled = true

	if len(pairs)%2 != 0 {
		panic("Pairs requires even number of arguments")
	}

	equals := "="
	if b.options.SpaceAroundFieldAssignment {
		equals = " = "
	}

	colon := ":"
	if b.options.SpaceAfterPairsSeparator {
		colon = ": "
	}

	parts := make([]string, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		k := formatValue(pairs[i])
		v := formatValue(pairs[i+1])
		parts = append(parts, k+colon+v)
	}

	separator := ";"
	if b.options.SpaceAfterListSeparator {
		separator = "; "
	}

	b.fields = append(b.fields, fmt.Sprintf("%s%s%s",
		name, equals, strings.Join(parts, separator)))
	return b
}

// String returns the built plainfields string
func (b *Builder) String() string {
	separator := ","
	if b.options.SpaceAfterFieldSeparator {
		separator = ", "
	}
	return strings.Join(b.fields, separator)
}

// NewBuilder creates a new plainfields builder with default options
func NewBuilder() *Builder {
	return NewBuilderWithOptions(BuilderDefaults())
}

// NewBuilderWithOptions creates a new builder with the specified options
func NewBuilderWithOptions(options BuilderOptions) *Builder {
	return &Builder{
		options: options,
	}
}
