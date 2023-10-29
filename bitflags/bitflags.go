package bitflags

// import "fmt"
const BITFIELD_SIZE = 64

type BitField uint64

// const (
//
//	F0 BitField = 1 << iota
//	F1
//	F2
//
// )
// const BITFIELD_ONE = BitField(1)

const BITFIELD_EMPTY = BitField(0)
const BITFIELD_FULL = BitField(0xFFFF_FFFF_FFFF_FFFF)

func (b BitField) Set(pos int) BitField    { return 1<<pos | b }
func (b BitField) Clear(pos int) BitField  { return ^(1 << pos) & b }
func (b BitField) Toggle(pos int) BitField { return 1<<pos ^ b }
func (b BitField) Has(pos int) bool        { return 1<<pos&b != 0 }

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
