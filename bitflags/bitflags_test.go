package bitflags

import (
	"testing"
)

// Check if bits are Set correctly.
func TestCorrectlySetsBit(t *testing.T) {
	b := BITFIELD_EMPTY

	b = Set(b, 2)
	if b != 0b100 {
		t.Fatalf("Expected bit 2 to be set, instead got %04b.\n", b)
	}

	b = Set(b, 0)
	if b != 0b101 {
		t.Fatalf("Expected bit 0 to be set, instead got %04b.\n", b)
	}
	b = Set(b, 3)
	if b != 0b1101 {
		t.Fatalf("Expected bit 4 to be set, instead got %04b.\n", b)
	}
	b = Set(b, BITFIELD_SIZE)
	if b != 0b1101 {
		t.Fatalf("Expected bits beyond bitfield size nevet to be set.\n")
	}
}

// Check if bits are Cleared correctly.
func TestCorrectlyClearsBit(t *testing.T) {
	b := BitField(0b1101)

	b = Clear(b, 2)
	if b != 0b1001 {
		t.Fatalf("Expected bit 2 to be clear, instead got %04b.\n", b)
	}

	b = Clear(b, 0)
	if b != 0b1000 {
		t.Fatalf("Expected bit 0 to be clear, instead got %04b.\n", b)
	}
	b = Clear(b, 3)
	if b != 0b0000 {
		t.Fatalf("Expected bit 3 to be clear, instead got %04b.\n", b)
	}
	b = Clear(b, 0)
	if b != 0b0000 {
		t.Fatalf("Expected bit 0 to be clear, instead got %04b.\n", b)
	}
}

// Check if bits are Toggled correctly.
func TestCorrectlyToggledBit(t *testing.T) {
	b := BitField(0b1101)

	// v := b.Toggle(2)
	b = Toggle(b, 2)
	if b != 0b1001 {
		t.Fatalf("Expected bit 2 to be clear, instead got %04b.\n", b)
	}
	b = Toggle(b, 2)
	if b != 0b1101 {
		t.Fatalf("Expected bit 2 to be set, instead got %04b.\n", b)
	}

	b = Toggle(b, 0)
	if b != 0b1100 {
		t.Fatalf("Expected bit 0 to be clear, instead got %04b.\n", b)
	}
	b = Toggle(b, 0)
	if b != 0b1101 {
		t.Fatalf("Expected bit 0 to be set, instead got %04b.\n", b)
	}

	b = Toggle(b, 3)
	if b != 0b0101 {
		t.Fatalf("Expected bit 3 to be clear, instead got %04b.\n", b)
	}
	b = Toggle(b, 3)
	if b != 0b1101 {
		t.Fatalf("Expected bit 0 to be set, instead got %04b.\n", b)
	}
}

// Check if bits are checked correctly.
func TestCorrectlyChecksdBit(t *testing.T) {
	b := BitField(0b1101)

	if !Has(b, 0) {
		t.Fatalf("Expected check bit 0 to be true.\n")
	}
	if Has(b, 1) {
		t.Fatalf("Expected check bit 1 to be false.\n")
	}
	if !Has(b, 2) {
		t.Fatalf("Expected check bit 2 to be true.\n")
	}
	if !Has(b, 3) {
		t.Fatalf("Expected check bit 3 to be true.\n")
	}
	if Has(b, BITFIELD_SIZE) {
		t.Fatalf("Expected check bit beyond bitfield size to be false.\n")
	}
}
