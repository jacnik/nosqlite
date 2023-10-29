package bitflags

type BitsBlock uint64

const (
	BitsBlockSize  = 64
	BitsBlockEmpty = BitsBlock(0)
	BitsBlockFull  = BitsBlock(0xFFFF_FFFF_FFFF_FFFF)
)

func (b BitsBlock) Set(pos uint) BitsBlock          { return 1<<pos | b }
func (b BitsBlock) Clear(pos uint) BitsBlock        { return ^(1 << pos) & b }
func (b BitsBlock) Toggle(pos uint) BitsBlock       { return 1<<pos ^ b }
func (b BitsBlock) Has(pos uint) bool               { return 1<<pos&b != 0 }
func (b BitsBlock) Union(o BitsBlock) BitsBlock     { return b | o }
func (b BitsBlock) Intersect(o BitsBlock) BitsBlock { return b & o }

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
*/
