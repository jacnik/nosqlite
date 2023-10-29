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

func Set(b BitField, pos int) BitField    { return 1<<pos | b }
func Clear(b BitField, pos int) BitField  { return ^(1 << pos) & b }
func Toggle(b BitField, pos int) BitField { return 1<<pos ^ b }
func Has(b BitField, pos int) bool        { return 1<<pos&b != 0 }

// func Set(b, flag BitField) BitField    { return b | flag }
// func Clear(b, flag BitField) BitField  { return b &^ flag }
// func Toggle(b, flag BitField) BitField { return b ^ flag }
// func Has(b, flag BitField) bool        { return b&flag != 0 }

// func main() {
// 	var b BitField
// 	b = Set(b, F0)
// 	b = Toggle(b, F2)
// 	for i, flag := range []BitField{F0, F1, F2} {
// 		fmt.Println(i, Has(b, flag))
// 	}
// }
