package bitflags

// import "fmt"
type BitsBlock uint64

const (
	BitsBlock_SIZE  = 64
	BitsBlock_EMPTY = BitsBlock(0)
	BitsBlock_FULL  = BitsBlock(0xFFFF_FFFF_FFFF_FFFF)
)

func (b BitsBlock) Set(pos int) BitsBlock           { return 1<<pos | b }
func (b BitsBlock) Clear(pos int) BitsBlock         { return ^(1 << pos) & b }
func (b BitsBlock) Toggle(pos int) BitsBlock        { return 1<<pos ^ b }
func (b BitsBlock) Has(pos int) bool                { return 1<<pos&b != 0 }
func (b BitsBlock) Union(o BitsBlock) BitsBlock     { return b | o }
func (b BitsBlock) Intersect(o BitsBlock) BitsBlock { return b & o }

// func Set(b, flag BitField) BitField    { return b | flag }
// func Clear(b, flag BitField) BitField  { return b &^ flag }
// func Toggle(b, flag BitField) BitField { return b ^ flag }
// func Has(b, flag BitField) bool        { return b&flag != 0 }

// type BitField interface {
// 	Set(pos int) BitField
// 	Clear(pos int) BitField
// 	Toggle(pos int) BitField
// 	Has(pos int) bool
// }

// https://itecnote.com/tecnote/go-how-would-you-set-and-clear-a-single-bit-in-go/
// http://graphics.stanford.edu/~seander/bithacks.html
