package plainfields

import (
	"fmt"
	"strings"
)

var (
	ErrOddNumberOfPairs         = fmt.Errorf("odd number of pairs")
	ErrOrderedFieldAfterLabeled = fmt.Errorf("ordered field after labeled field")
	ErrInvalidFieldName         = fmt.Errorf("invalid field name")
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
	hasLabeled bool   // Track if we've seen any labeled fields.
	nextLabel  string // Track the next label to be used.
	err        error  // Track the last error that occurred.
}

// setError sets the last error that occurred.
func (b *Builder) setError(err error) *Builder {
	if b.err == nil {
		b.err = err
	}
	return b
}

// addRaw adds a raw value to the builder.
func (b *Builder) addRaw(value string) *Builder {
	b.nextLabel = ""
	b.fields = append(b.fields, value)
	return b
}

// add adds a new value to the builder.
func (b *Builder) add(value string) *Builder {
	if b.err != nil {
		return b
	}

	if b.nextLabel != "" {
		separator := "="
		if b.options.SpaceAroundFieldAssignment {
			separator = " = "
		}
		value = fmt.Sprintf("%s%s%s", b.nextLabel, separator, value)

		b.hasLabeled = true

	} else if b.hasLabeled {
		// If we have seen a labeled field, we cannot add an ordered value.
		return b.setError(ErrOrderedFieldAfterLabeled)
	}

	return b.addRaw(value)
}

// Err returns the last error that occurred if any.
func (b *Builder) Err() error {
	return b.err
}

// Value adds an ordered value to the builder.
func (b *Builder) Value(value any) *Builder {
	return b.add(b.options.formatValue(value))
}

// List adds a [name=]value1;value2;... field.
func (b *Builder) List(values ...any) *Builder {
	items := make([]string, len(values))
	for i, v := range values {
		items[i] = b.options.formatValue(v)
	}

	separator := ";"
	if b.options.SpaceAfterListSeparator {
		separator = "; "
	}

	return b.add(strings.Join(items, separator))
}

// Dict adds a [name=]key1:value1;key2:value2;... field.
func (b *Builder) Dict(pairs ...any) *Builder {
	if len(pairs)%2 != 0 {
		return b.setError(ErrOddNumberOfPairs)
	}

	colon := ":"
	if b.options.SpaceAfterPairsSeparator {
		colon = ": "
	}

	items := make([]string, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		k := b.options.formatValue(pairs[i])
		v := b.options.formatValue(pairs[i+1])
		items = append(items, k+colon+v)
	}

	separator := ";"
	if b.options.SpaceAfterListSeparator {
		separator = "; "
	}
	return b.add(strings.Join(items, separator))
}

// Enable adds a boolean field with ^ prefix.
func (b *Builder) Enable(name string) *Builder {
	b.hasLabeled = true
	return b.addRaw("^" + name)
}

// Disable adds a boolean field with ! prefix.
func (b *Builder) Disable(name string) *Builder {
	b.hasLabeled = true
	return b.addRaw("!" + name)
}

// Boolean adds a boolean field with ^ or ! prefix.
func (b *Builder) Boolean(name string, value bool) *Builder {
	if value {
		return b.Enable(name)
	}
	return b.Disable(name)
}

// Label sets the name of the field for the next value.
func (b *Builder) Label(name string) *Builder {
	if NeedsQuoting(name) {
		return b.setError(fmt.Errorf("%q: %w", name, ErrInvalidFieldName))
	}
	b.nextLabel = name
	return b
}

// Labeled adds a name=value field
func (b *Builder) Labeled(name string, value any) *Builder {
	return b.Label(name).Value(value)
}

// LabeledList adds a [name=]value1;value2[;...] field
func (b *Builder) LabeledList(name string, values ...any) *Builder {
	return b.Label(name).List(values...)
}

// LabeledDict adds a [name=]key1:value1;key2:value2[;...] field
func (b *Builder) LabeledDict(name string, pairs ...any) *Builder {
	return b.Label(name).Dict(pairs...)
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
