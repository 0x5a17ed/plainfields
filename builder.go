package plainfields

import (
	"fmt"
	"strings"
)

var (
	ErrOddNumberOfPairs         = fmt.Errorf("odd number of pairs")
	ErrOrderedFieldAfterLabeled = fmt.Errorf("ordered field after labeled field")
)

// NeedsQuoting returns true if the given string needs quotes.
func NeedsQuoting(s string) bool {
	return s == "" ||
		strings.ContainsAny(s, ` ,;:=\`) ||
		s == "true" || s == "false" || s == "nil"
}

// BuilderOptions controls Builder formatting behavior.
type BuilderOptions struct {
	// AlwaysQuoteStrings forces all strings to be quoted.
	AlwaysQuoteStrings bool

	// SpaceAfterFieldSeparator adds a space after the field separator `,`.
	SpaceAfterFieldSeparator bool

	// SpaceAfterListSeparator adds a space after the list separator `;`.
	SpaceAfterListSeparator bool

	// SpaceAfterPairsSeparator adds a space after the map key-value separator `:`.
	SpaceAfterPairsSeparator bool

	// SpaceAroundFieldAssignment adds a space around the field assignment `=`.
	SpaceAroundFieldAssignment bool
}

// formatValue returns a string representation suitable for plainfields.
func (opt BuilderOptions) formatValue(v any) string {
	switch val := v.(type) {
	case string:
		if opt.AlwaysQuoteStrings || NeedsQuoting(val) {
			return fmt.Sprintf("%q", val)
		}
		return val
	case nil:
		return "nil"
	default:
		return fmt.Sprint(v)
	}
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
	hasLabeled bool  // Track if we've seen any labeled fields.
	err        error // Track the last error that occurred.
}

// setError sets the last error that occurred.
func (b *Builder) setError(err error) *Builder {
	if b.err == nil {
		b.err = err
	}
	return b
}

// add adds a new field to the builder.
func (b *Builder) add(field string) *Builder {
	if b.err != nil {
		return b
	}
	b.fields = append(b.fields, field)
	return b
}

// Err returns the last error that occurred if any.
func (b *Builder) Err() error {
	return b.err
}

// Ordered adds an ordered value to the builder.
func (b *Builder) Ordered(value any) *Builder {
	if b.hasLabeled {
		return b.setError(ErrOrderedFieldAfterLabeled)
	}
	return b.add(b.options.formatValue(value))
}

// Enable adds a boolean field with ^ prefix.
func (b *Builder) Enable(name string) *Builder {
	b.hasLabeled = true
	return b.add("^" + name)
}

// Disable adds a boolean field with ! prefix.
func (b *Builder) Disable(name string) *Builder {
	b.hasLabeled = true
	return b.add("!" + name)
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
	return b.add(fmt.Sprintf("%s%s%v", name, equals, b.options.formatValue(value)))
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
		parts[i] = b.options.formatValue(v)
	}

	separator := ";"
	if b.options.SpaceAfterListSeparator {
		separator = "; "
	}

	return b.add(fmt.Sprintf("%s%s%s", name, equals, strings.Join(parts, separator)))
}

// Pairs adds a name=key1:value1;key2:value2 field
func (b *Builder) Pairs(name string, pairs ...any) *Builder {
	b.hasLabeled = true

	if len(pairs)%2 != 0 {
		return b.setError(ErrOddNumberOfPairs)
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
		k := b.options.formatValue(pairs[i])
		v := b.options.formatValue(pairs[i+1])
		parts = append(parts, k+colon+v)
	}

	separator := ";"
	if b.options.SpaceAfterListSeparator {
		separator = "; "
	}
	return b.add(fmt.Sprintf("%s%s%s", name, equals, strings.Join(parts, separator)))
}

// String returns the built plainfields string
func (b *Builder) String() string {
	if b.err != nil {
		return ""
	}

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
