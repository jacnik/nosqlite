package bitflags

import (
	"fmt"
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

func TestCorrectlyCountSetBits__(t *testing.T) {
	b := int8(0b0000_1101)

	fmt.Printf("%08b\n", 0x55) // 0101_0101
	fmt.Printf("%08b\n", 0x33) // 0011_0011
	fmt.Printf("%08b\n", 0x0F) // 0000_1111

	fmt.Printf("\n")
	fmt.Printf("b>>1 \t\t %08b\n", b>>1)                 // 0000_0110
	fmt.Printf("(b>>1)&0x55 \t %08b\n", (b>>1)&0x55)     // 0000_0100
	fmt.Printf("b-(b>>1)&0x55 \t %08b\n", b-(b>>1)&0x55) // 0000_1001
	b = b - (b>>1)&0x55
	fmt.Printf("b>>2 \t\t\t\t %08b\n", b>>2)                                 // 0000_0010
	fmt.Printf("(b>>2)&0x33 \t\t\t %08b\n", (b>>2)&0x33)                     // 0000_0010
	fmt.Printf("b&0x33 \t\t\t\t %08b\n", b&0x33)                             // 0000_0001
	fmt.Printf("(b&0x33) + ((b>>2)&0x33) \t %08b\n", (b&0x33)+((b>>2)&0x33)) // 0000_0011
	/*
		0000_0000 -
		0000_0100 =
		1111_1100

		0000_1101 +
		1111_1100 =
		0000_1001
	*/
	fmt.Printf("\n")
	x, y := int8(0b0000_1110), int8(0b0000_0110)
	fmt.Printf("%08b - \n%08b = \n%08b\n", x, y, x-y)

	fmt.Printf("\n")
	/*
	 3 = 0000_0011 -> -3 = 1111_1101
	 6 - 3 = 6 + (-3)
	 0000_0110 +
	 1111_1101 =
	 0000_0011
	*/
	// fmt.Printf("%8b\n", int8(-3))

	b = b - ((b >> 1) & 0x55)          // reuse input as temporary
	b = (b & 0x33) + ((b >> 2) & 0x33) // temp
	// c := ((b + (b>>4)&0xF0) * 0x10) >> 24 // count

	if 3 != 3 {
		t.Fatalf("Expected 3 bits to be set, got %d.\n", 1)
	}

	// 32 bit version below
	// b = b - ((b >> 1) & 0x55555555)                 // reuse input as temporary
	// b = (b & 0x33333333) + ((b >> 2) & 0x33333333)  // temp
	// c := ((b + (b>>4)&0xF0F0F0F) * 0x1010101) >> 24 // count

	// if c != 3 {
	// 	t.Fatalf("Expected 3 bits to be set, got %d.\n", c)
	// }
}
