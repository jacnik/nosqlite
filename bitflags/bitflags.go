package bitflags

type BitsBlock uint64

const (
	BitsBlockSize  = 64
	BitsBlockEmpty = BitsBlock(0)
	BitsBlockFull  = BitsBlock(0xFFFF_FFFF_FFFF_FFFF)

	m1  = BitsBlock(0x5555555555555555) //binary: 0101...
	m2  = BitsBlock(0x3333333333333333) //binary: 00110011..
	m4  = BitsBlock(0x0f0f0f0f0f0f0f0f) //binary:  4 zeros,  4 ones ...
	m8  = BitsBlock(0x00ff00ff00ff00ff) //binary:  8 zeros,  8 ones ...
	m16 = BitsBlock(0x0000ffff0000ffff) //binary: 16 zeros, 16 ones ...
	m32 = BitsBlock(0x00000000ffffffff) //binary: 32 zeros, 32 ones
	h01 = BitsBlock(0x0101010101010101) //the sum of 256 to the power of 0,1,2,3...
)

func (b BitsBlock) Set(pos uint) BitsBlock          { return 1<<pos | b }
func (b BitsBlock) Clear(pos uint) BitsBlock        { return ^(1 << pos) & b }
func (b BitsBlock) Toggle(pos uint) BitsBlock       { return 1<<pos ^ b }
func (b BitsBlock) Has(pos uint) bool               { return 1<<pos&b != 0 }
func (b BitsBlock) Union(o BitsBlock) BitsBlock     { return b | o }
func (b BitsBlock) Intersect(o BitsBlock) BitsBlock { return b & o }
func (b BitsBlock) Popcount() uint {
	b -= (b >> 1) & m1             //put count of each 2 bits into those 2 bits
	b = (b & m2) + ((b >> 2) & m2) //put count of each 4 bits into those 4 bits
	b = (b + (b >> 4)) & m4        //put count of each 8 bits into those 8 bits
	return uint((b * h01) >> 56)   //returns left 8 bits of x + (x<<8) + (x<<16) + (x<<24) + ...
	// https://en.wikipedia.org/wiki/Hamming_weight
}
func (b BitsBlock) Traverse() <-chan uint {
	res := make(chan uint, BitsBlockSize/2)

	go func() {
		for i := uint(0); b != 0; i, b = i+1, b>>1 {
			if b&1 == 1 {
				res <- i
			}
		}
		close(res)
	}()

	return res
}

/* Return position (0, 1, ...) of rightmost (least-significant) one bit in n.
 *
 * This code uses a 32-bit version of algorithm to find the rightmost
 * one bit in Knuth, _The Art of Computer Programming_, volume 4A
 * (draft fascicle), section 7.1.3, "Bitwise tricks and
 * techniques."
 *
 * Assumes n has a 1 bit, i.e. n != 0
 *
 */
// TODO 32 bit alternative to lsb
//  func rightone32(n uint32) uint32 {
// 	a := uint32(0x05f66a47) /* magic number, found by brute force */
// 	decode := [32]uint32{0, 1, 2, 26, 23, 3, 15, 27, 24, 21, 19, 4, 12, 16, 28, 6, 31, 25, 22, 14, 20, 18, 11, 5, 30, 13, 17, 10, 29, 9, 8, 7}
// 	n = a * (n & (-n))
// 	return decode[n>>27]
// }
func (b BitsBlock) Lsb() BitsBlock      { return b & -b }      // Clear all except lowest set bit
func (b BitsBlock) ClearLsb() BitsBlock { return b ^ b.Lsb() } // Clear lowest set bit

const (
	BitFlagsCap = BitsBlockSize * BitsBlockSize
)

type BitFlags struct { // like BitBlock but can hold up ints in [0, 64*64) range
	activeMask BitsBlock
	blocks     []BitsBlock
}

func divmod(numerator, denominator uint) (quotient, remainder uint) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

func (b *BitFlags) resizeBlocks(blockIndex uint) (newIndex uint) {
	if !b.activeMask.Has(blockIndex) {
		b.blocks = append(b.blocks, BitsBlockEmpty)
		b.activeMask = b.activeMask.Set(blockIndex)
	}
	return uint(len(b.blocks) - 1)
}
func (b *BitFlags) Set(positions ...uint) { // TODO Set assumes increasing arguments order
	for _, pos := range positions {
		blockMultiplier, blockPos := divmod(pos, BitsBlockSize)
		blockIndex := b.resizeBlocks(blockMultiplier)
		b.blocks[blockIndex] = b.blocks[blockIndex].Set(blockPos)
	}
}
func (b BitFlags) Traverse() <-chan uint {
	sizeGuess := b.activeMask.Popcount() * 32
	res := make(chan uint, sizeGuess)

	go func() {
		blockIndex := 0
		for blockMultiplier := range b.activeMask.Traverse() {
			for blockPos := range b.blocks[blockIndex].Traverse() {
				res <- blockMultiplier*BitsBlockSize + blockPos
			}
			blockIndex++
		}
		close(res)
	}()

	return res
}

// func (b BitFlags) Traverse() []uint {
// 	sizeGuess := b.activeMask.Popcount() * 32
// 	res := make([]uint, 0, sizeGuess)

// 	mask := b.activeMask
// 	for blockIndex := 0; mask != 0; blockIndex, mask = blockIndex+1, mask>>1 {
// 		if mask&1 == 1 {
// 			block := b.blocks[blockIndex]
// 			for blockPos := 0; block != 0; blockPos, block = blockPos+1, block>>1 {
// 				if block&1 == 1 {
// 					res = append(res, uint(blockIndex)*BitsBlockSize+uint(blockPos))
// 				}
// 			}
// 		}
// 	}

// 	return res
// }

// https://itecnote.com/tecnote/go-how-would-you-set-and-clear-a-single-bit-in-go/
// http://graphics.stanford.edu/~seander/bithacks.html

/*
B[0] = 0x55555555 = 01010101 01010101 01010101 01010101
B[1] = 0x33333333 = 00110011 00110011 00110011 00110011
B[2] = 0x0F0F0F0F = 00001111 00001111 00001111 00001111
B[3] = 0x00FF00FF = 00000000 11111111 00000000 11111111
B[4] = 0x0000FFFF = 00000000 00000000 11111111 11111111

 The best method for counting bits in a 32-bit integer v is the following:

v = v - ((v >> 1) & 0x55555555);                    // reuse input as temporary
v = (v & 0x33333333) + ((v >> 2) & 0x33333333);     // temp
c = ((v + (v >> 4) & 0xF0F0F0F) * 0x1010101) >> 24; // count

The best bit counting method takes only 12 operations, which is the same as the lookup-table method, but avoids the memory and potential cache misses of a table. It is a hybrid between the purely parallel method above and the earlier methods using multiplies (in the section on counting bits with 64-bit instructions), though it doesn't use 64-bit instructions. The counts of bits set in the bytes is done in parallel, and the sum total of the bits set in the bytes is computed by multiplying by 0x1010101 and shifting right 24 bits.

A generalization of the best bit counting method to integers of bit-widths upto 128 (parameterized by type T) is this:

v = v - ((v >> 1) & (T)~(T)0/3);                           // temp
v = (v & (T)~(T)0/15*3) + ((v >> 2) & (T)~(T)0/15*3);      // temp
v = (v + (v >> 4)) & (T)~(T)0/255*15;                      // temp
c = (T)(v * ((T)~(T)0/255)) >> (sizeof(T) - 1) * CHAR_BIT; // count

See Ian Ashdown's nice newsgroup post for more information on counting the number of bits set (also known as sideways addition). The best bit counting method was brought to my attention on October 5, 2005 by Andrew Shapira; he found it in pages 187-188 of Software Optimization Guide for AMD Athlon™ 64 and Opteron™ Processors. Charlie Gordon suggested a way to shave off one operation from the purely parallel version on December 14, 2005, and Don Clugston trimmed three more from it on December 30, 2005. I made a typo with Don's suggestion that Eric Cole spotted on January 8, 2006. Eric later suggested the arbitrary bit-width generalization to the best method on November 17, 2006. On April 5, 2007, Al Williams observed that I had a line of dead code at the top of the first method.
*/
