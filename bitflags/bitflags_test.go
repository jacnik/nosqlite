package bitflags

import (
	"testing"
)

// Check if bits are Set correctly.
func TestCorrectlySetsBit(t *testing.T) {
	b := BitField(1)
	b = Set(b, 2)

	if !Has(b, 4) {
		t.Fatalf("Expected bit at position 3 to be set, instead got %04b.\n", b)
	}
}
