package bitflags

import (
	"testing"
)

// Check if bits are Set correctly.
func TestCorrectlySetsBit(t *testing.T) {
	b := BitsBlockEmpty

	b = b.Set(2)
	if b != 0b100 {
		t.Fatalf("Expected bit 2 to be set, instead got %04b.\n", b)
	}
	b = b.Set(0)
	if b != 0b101 {
		t.Fatalf("Expected bit 0 to be set, instead got %04b.\n", b)
	}
	b = b.Set(3)
	if b != 0b1101 {
		t.Fatalf("Expected bit 4 to be set, instead got %04b.\n", b)
	}
	b = b.Set(BitsBlockSize)
	if b != 0b1101 {
		t.Fatalf("Expected bits beyond bitfield size nevet to be set.\n")
	}
}

// Check if bits are Cleared correctly.
func TestCorrectlyClearsBit(t *testing.T) {
	b := BitsBlock(0b1101)

	b = b.Clear(2)
	if b != 0b1001 {
		t.Fatalf("Expected bit 2 to be clear, instead got %04b.\n", b)
	}
	b = b.Clear(0)
	if b != 0b1000 {
		t.Fatalf("Expected bit 0 to be clear, instead got %04b.\n", b)
	}
	b = b.Clear(3)
	if b != 0b0000 {
		t.Fatalf("Expected bit 3 to be clear, instead got %04b.\n", b)
	}
	b = b.Clear(0)
	if b != 0b0000 {
		t.Fatalf("Expected bit 0 to be clear, instead got %04b.\n", b)
	}
}

// Check if bits are Toggled correctly.
func TestCorrectlyToggledBit(t *testing.T) {
	b := BitsBlock(0b1101)

	b = b.Toggle(2)
	if b != 0b1001 {
		t.Fatalf("Expected bit 2 to be clear, instead got %04b.\n", b)
	}
	b = b.Toggle(2)
	if b != 0b1101 {
		t.Fatalf("Expected bit 2 to be set, instead got %04b.\n", b)
	}

	b = b.Toggle(0)
	if b != 0b1100 {
		t.Fatalf("Expected bit 0 to be clear, instead got %04b.\n", b)
	}
	b = b.Toggle(0)
	if b != 0b1101 {
		t.Fatalf("Expected bit 0 to be set, instead got %04b.\n", b)
	}

	b = b.Toggle(3)
	if b != 0b0101 {
		t.Fatalf("Expected bit 3 to be clear, instead got %04b.\n", b)
	}
	b = b.Toggle(3)
	if b != 0b1101 {
		t.Fatalf("Expected bit 0 to be set, instead got %04b.\n", b)
	}
}

// Check if bits are checked correctly.
func TestCorrectlyChecksBit(t *testing.T) {
	b := BitsBlock(0b1101)

	if !b.Has(0) {
		t.Fatalf("Expected check bit 0 to be true.\n")
	}
	if b.Has(1) {
		t.Fatalf("Expected check bit 1 to be false.\n")
	}
	if !b.Has(2) {
		t.Fatalf("Expected check bit 2 to be true.\n")
	}
	if !b.Has(3) {
		t.Fatalf("Expected check bit 3 to be true.\n")
	}
	if b.Has(BitsBlockSize) {
		t.Fatalf("Expected check bit beyond bitfield size to be false.\n")
	}
}

// Check if bit flags are unioned correctly.
func TestCorrectlyUnionBitFlags(t *testing.T) {
	b := BitsBlock(0b1101)

	u := b.Union(0b0110)
	if u != 0b1111 {
		t.Fatalf("Expected union to be '1111', got '%04b'.\n", u)
	}
}

// Check if bit flags are intersected correctly.
func TestCorrectlyIntersectBitFlags(t *testing.T) {
	b := BitsBlock(0b1101)

	u := b.Intersect(0b0110)
	if u != 0b0100 {
		t.Fatalf("Expected intersetion to be '0100', got '%04b'.\n", u)
	}
}

func TestCorrectlyCountSetBits(t *testing.T) {
	assert := func(b BitsBlock, popcount uint) {
		if b.Popcount() != popcount {
			t.Fatalf("Expected %d bits to be set in %b, got %d.\n", popcount, b, b.Popcount())
		}
	}
	assert(BitsBlock(0b0), 0)
	assert(BitsBlock(0b0100), 1)
	assert(BitsBlock(0b1001), 2)
	assert(BitsBlock(0b1101), 3)
	assert(BitsBlock(0b1111), 4)
	assert(BitsBlock(0b0110_1111), 6)
	assert(BitsBlock(0b1101_0110_1111), 9)
	assert(BitsBlockFull, BitsBlockSize)
}
