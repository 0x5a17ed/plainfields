package plainfields

import (
	"fmt"
)

// Position holds the offset in the input along with line and column numbers.
type Position struct {
	Offset int // Byte Offset into the input.
	Column int // Column number (starting at 1).
}

func (p Position) String() string {
	return fmt.Sprintf("Col %d (Offset %d)", p.Column, p.Offset)
}
