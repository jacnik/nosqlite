package bitflags

// import "fmt"
const BITFIELD_SIZE = 64

type BitFlags uint64

// const (
//
//	F0 BitField = 1 << iota
//	F1
//	F2
//
// )
// const BITFIELD_ONE = BitField(1)

const BITFIELD_EMPTY = BitFlags(0)
const BITFIELD_FULL = BitFlags(0xFFFF_FFFF_FFFF_FFFF)

func (b BitFlags) Set(pos int) BitFlags          { return 1<<pos | b }
func (b BitFlags) Clear(pos int) BitFlags        { return ^(1 << pos) & b }
func (b BitFlags) Toggle(pos int) BitFlags       { return 1<<pos ^ b }
func (b BitFlags) Has(pos int) bool              { return 1<<pos&b != 0 }
func (b BitFlags) Union(o BitFlags) BitFlags     { return b | o }
func (b BitFlags) Intersect(o BitFlags) BitFlags { return b & o }

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
