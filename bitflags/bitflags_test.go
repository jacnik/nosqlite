package bitflags

import (
	"testing"
)

// BitsBlock: Check if bits are Set correctly.
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

// BitsBlock: Check if bits are Cleared correctly.
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

// BitsBlock: Check if bits are Toggled correctly.
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

// BitsBlock: Check if bits are checked correctly.
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

// BitsBlock: Check if bit flags are unioned correctly.
func TestCorrectlyUnionBitFlags(t *testing.T) {
	b := BitsBlock(0b1101)

	u := b.Union(0b0110)
	if u != 0b1111 {
		t.Fatalf("Expected union to be '1111', got '%04b'.\n", u)
	}
}

// BitsBlock: Check if bit flags are intersected correctly.
func TestCorrectlyIntersectBitFlags(t *testing.T) {
	b := BitsBlock(0b1101)

	u := b.Intersect(0b0110)
	if u != 0b0100 {
		t.Fatalf("Expected intersetion to be '0100', got '%04b'.\n", u)
	}
}

// BitsBlock: Check if correctly counts set bits.
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

// BitsBlock: Check if correctly returns lowest set bits.
func TestCorrectlyReturnsLowestSetBit(t *testing.T) {
	assertLsb := func(b, expected BitsBlock) {
		if b.Lsb() != expected {
			t.Fatalf("Expected lowest set bit to be: %04b, got: %04b.\n", expected, b.Lsb())
		}
	}
	assertLsb(0b0, 0b0)
	assertLsb(0b1101, 0b1)
	assertLsb(0b1010, 0b10)
	assertLsb(0b1100, 0b100)
	assertLsb(0b1000, 0b1000)
}

// BitsBlock: Check if correctly clears lowest set bits.
func TestCorrectlyClearsLowestSetBit(t *testing.T) {
	assertClearLsb := func(b, expected BitsBlock) {
		if b.ClearLsb() != expected {
			t.Fatalf("Expected lowest set bit to be: %04b, got: %04b.\n", expected, b.ClearLsb())
		}
	}
	assertClearLsb(0b0, 0b0)
	assertClearLsb(0b1000, 0b0)
	assertClearLsb(0b1101, 0b1100)
	assertClearLsb(0b1010, 0b1000)
	assertClearLsb(0b1100, 0b1000)
}

func compareRanges(a <-chan uint, b []uint) bool {
	i := 0
	for v := range a {
		if b[i] != v {
			return false
		}
		i++
	}
	if i < len(b) {
		return false
	}
	return true
}

// BitFlags: Check if correctly counts set bits.
func TestSetAndTraverseBitFlags(t *testing.T) {
	assert := func(f BitFlags, expected []uint) {
		actual := f.Traverse()
		if !compareRanges(actual, expected) {
			t.Fatalf("Expected range different than actual:\n%v\n", expected)
		}
	}

	o := BitFlags{}
	assert(o, []uint{})

	f := BitFlags{}
	f.Set(1, 64, 129, 64*3+18, 64*3+20, 64*64-1)
	assert(f, []uint{1, 64, 129, 64*3 + 18, 64*3 + 20, 64*64 - 1})

	w := BitFlags{}
	allIndexes := make([]uint, 64*64)
	for i := uint(0); i < 64*64; i++ {
		allIndexes[i] = i
		w.Set(i)
	}
	assert(w, allIndexes)
}

func compareBitsBlocks(a, b []BitsBlock) bool {
	i := 0
	for _, v := range a {
		if b[i] != v {
			return false
		}
		i++
	}
	if i < len(b) {
		return false
	}
	return true
}

func compareBitFlags(a, b BitFlags) bool {
	if a.activeMask != b.activeMask {
		return false
	}
	return compareBitsBlocks(a.blocks, b.blocks)
}

// BitFlags: Check if correctly union bits flags.
func TestUnionBitFlags(t *testing.T) {
	assert := func(expected, actual BitFlags) {
		if !compareBitFlags(expected, actual) {
			t.Fatalf("Expected BitFlags different than actual:\n%v\n%v\n", expected, actual)
		}
	}

	assert(BitFlags{}.Union(BitFlags{}), BitFlags{})

	a, b := BitFlags{}, BitFlags{}
	a.Set(0)
	assert(a.Union(b), BitFlags{activeMask: 1, blocks: []BitsBlock{1}})

	a, b = BitFlags{}, BitFlags{}
	a.Set(1)
	assert(a.Union(b), BitFlags{activeMask: 1, blocks: []BitsBlock{2}})

	a, b = BitFlags{}, BitFlags{}
	a.Set(2)
	b.Set(3)
	assert(a.Union(b), BitFlags{activeMask: 1, blocks: []BitsBlock{0b1100}})

	a, b = BitFlags{}, BitFlags{}
	a.Set(3)
	b.Set(3)
	assert(a.Union(b), BitFlags{activeMask: 1, blocks: []BitsBlock{0b1000}})

	a, b = BitFlags{}, BitFlags{}
	a.Set(2, 64, 64*2+1, 64*3+18, 64*3+20, 64*64-1)
	b.Set(3, 63, 64*4+1, 64*5+2)
	assert(a.Union(b), BitFlags{
		activeMask: 0b10000000_00000000_00000000_00000000_00000000_00000000_00000000_00111111, // 64*0, 64*1, 64*2, 64*3, 64*4, 64*5, 64*64
		blocks: []BitsBlock{
			0b10000000_00000000_00000000_00000000_00000000_00000000_00000000_00001100, // 2, 3, 63
			0b0001,                       // 64
			0b0010,                       // 64*2+1,
			0b00010100_00000000_00000000, // 64*3+18, 64*3+20
			0b0010,                       // 64*4+1,
			0b0100,                       // 64*5+2
			0b10000000_00000000_00000000_00000000_00000000_00000000_00000000_00000000, // 64*64-1
		}})
}

// func TestDivideBy32(t *testing.T) {
// testDivBy32 := func() {
// 	// Dividing by 32 is the same as bit shift left by 5
// 	for n := 0; n <= 0b11111111_11111111_11111111_11111111; n++ {
// 		if n/32 != n>>5 {
// 			fmt.Printf("Difference for n = %d\n%08b\n%08b\n", n, n/32, n>>5)
// 			t.Fatalf("Difference for n = %d\n%08b\n%08b\n", n, n/32, n>>5)
// 		}
// 	}
// }
// testDivBy32()

// testModBy32 := func() {
// 	// Mod by 32 is the same as binary and with 31
// 	m := 0b11111 // 31 0x1F
// 	for n := 0; n <= 0b11111111_11111111_11111111_11111111; n++ {
// 		v := n & m
// 		if v != n%32 {
// 			t.Fatalf("Difference for n = %d\n%08b\n%08b\n", n, n%32, v)
// 		}
// 	}
// }
// testModBy32()
// }
