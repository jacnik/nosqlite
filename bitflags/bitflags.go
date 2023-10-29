package bitflags

// import "fmt"
const BITFIELD_SIZE = 64

type BitField uint64

// const (
// 	F0 BitField = 1 << iota
// 	F1
// 	F2
// )

func Set(b, flag BitField) BitField    { return b | flag }
func Clear(b, flag BitField) BitField  { return b &^ flag }
func Toggle(b, flag BitField) BitField { return b ^ flag }
func Has(b, flag BitField) bool        { return b&flag != 0 }

// func main() {
// 	var b BitField
// 	b = Set(b, F0)
// 	b = Toggle(b, F2)
// 	for i, flag := range []BitField{F0, F1, F2} {
// 		fmt.Println(i, Has(b, flag))
// 	}
// }
