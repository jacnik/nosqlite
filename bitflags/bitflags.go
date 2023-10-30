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

const (
	BitFlagsCap = BitsBlockSize * BitsBlockSize
)

type BitFlags struct {
	active BitsBlock
	blocks []BitsBlock
}

func divmod(numerator, denominator uint) (quotient, remainder uint) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

func (b *BitFlags) resizeBlocks() {
	if len(b.blocks) == 0 {
		b.blocks = make([]BitsBlock, BitsBlockSize)
	}
}

func (b *BitFlags) Set(pos uint) {
	blockIndex, blockPos := divmod(pos, BitsBlockSize)
	b.resizeBlocks()

	b.active = b.active.Set(blockIndex)
	b.blocks[blockIndex] = b.blocks[blockIndex].Set(blockPos)
}

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
