package main

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"slices"
	"strconv"

	"github.com/jacnik/bitflags"
)

type size_t uint32
type separator byte

const (
	NUL separator = '\x00'
	SOH separator = '\x01'
	STX separator = '\x02'
	ETX separator = '\x03'
	LF  separator = '\x0A'
	FF  separator = '\x0C'
	DEL separator = '\x7f'
)

type IndexEntryType byte

const (
	FloatType IndexEntryType = 'f'
	StrType   IndexEntryType = 's'
	NullType  IndexEntryType = 'n'
)

type ValueRefs struct {
	value interface{}
	refs  []size_t
}

func (v ValueRefs) String() string {
	return fmt.Sprintf("%v %v", v.value, v.refs)
}

type IndexEntry struct {
	key       string
	valueType IndexEntryType
	values    []ValueRefs
}

type IndexT []IndexEntry

/* Types used when aggregating json */
/* **** */
type aggregateKeyT struct {
	key       string
	valueType IndexEntryType
}
type aggregateValueT interface{}
type aggregateFileRefT []size_t
type aggregateT map[aggregateKeyT]map[aggregateValueT]aggregateFileRefT

type flattenJsonT map[aggregateKeyT]aggregateValueT

/* **** */

func check(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func listDir(path string) []string {
	files, err := os.ReadDir(path)
	check(err)

	names := make([]string, 0, 16)
	for _, entry := range files {
		if !entry.IsDir() && entry.Name() != "INDEX" {
			names = append(names, entry.Name())
		}
	}
	return names
}

func readFile(path string) []byte {
	jsonFile, err := os.Open(path)
	check(err)
	defer jsonFile.Close()

	bytes, err := io.ReadAll(jsonFile)
	check(err)

	return bytes
}

func parseJson(bytes []byte) interface{} {
	var result interface{}
	json.Unmarshal(bytes, &result)
	return result
}

func flattenJsonMap(flatten flattenJsonT, prefix string, jMap map[string]interface{}) {
	for key, jItem := range jMap {
		flattenWithPrefix(flatten, prefix+"/"+key, jItem)
	}
}

func flattenJsonArr(flatten flattenJsonT, prefix string, jArr []interface{}) {
	for i, jItem := range jArr {
		flattenWithPrefix(flatten, prefix+"/"+strconv.Itoa(i), jItem)
	}
}

func flattenWithPrefix(flatten flattenJsonT, prefix string, unflatten interface{}) {
	switch v := unflatten.(type) {
	case map[string]interface{}:
		flattenJsonMap(flatten, prefix, v)
	case []interface{}:
		flattenJsonArr(flatten, prefix, v)
	case string:
		flatten[aggregateKeyT{prefix, StrType}] = v
	case float64:
		flatten[aggregateKeyT{prefix, FloatType}] = v
	default:
		flatten[aggregateKeyT{prefix, NullType}] = v
	}
}

func flattenJson(unflatten interface{}) flattenJsonT {
	flatten := make(flattenJsonT)
	flattenWithPrefix(flatten, "", unflatten)
	return flatten
}

func aggregateJson(agg aggregateT, flatten flattenJsonT, fileIdx size_t) {
	for aggregateKey, aggregateValue := range flatten {
		fileRefsMap, hasFileRefsMap := agg[aggregateKey]
		if !hasFileRefsMap {
			fileRefsMap = make(map[aggregateValueT]aggregateFileRefT)
			fileRefsMap[aggregateValue] = []size_t{fileIdx}
			agg[aggregateKey] = fileRefsMap
		} else {
			fileRefsMap[aggregateValue] = append(fileRefsMap[aggregateValue], fileIdx)
		}
	}
}

func sortValues(values map[aggregateValueT]aggregateFileRefT, valueType IndexEntryType) []ValueRefs {
	aggregateValueCmp := func(a, b aggregateValueT) int {
		/* Should return a negative number when a < b,
		** a positive number when a > b
		** and zero when a == b. */
		switch valueType {
		case FloatType:
			return cmp.Compare(a.(float64), b.(float64))
		case StrType:
			return cmp.Compare(a.(string), b.(string))
		case NullType:
			return 0
		}
		message := fmt.Sprintf("Unknown type %c\n", valueType)
		panic(message)
	}

	keys := make([]aggregateValueT, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}

	slices.SortFunc(keys, aggregateValueCmp)

	sortedRefs := make([]ValueRefs, 0, len(keys))
	for _, key := range keys {
		sortedRefs = append(sortedRefs, ValueRefs{key, values[key]})
	}
	return sortedRefs
}

func valueWithTypeCmp[T cmp.Ordered](valueA, valueB T, typeA, typeB IndexEntryType) int {
	valueCmp := cmp.Compare(valueA, valueB)
	if valueCmp != 0 {
		return valueCmp
	}

	return cmp.Compare(typeA, typeB)
}

func sortedAggregateKeys(mapToSort aggregateT) []aggregateKeyT {
	aggregateKeyCmp := func(x, y aggregateKeyT) int {
		return valueWithTypeCmp(x.key, y.key, x.valueType, y.valueType)
	}
	keys := make([]aggregateKeyT, 0, len(mapToSort))
	for p := range mapToSort {
		keys = append(keys, p)
	}

	slices.SortFunc(keys, aggregateKeyCmp)
	return keys
}

func indexAgregate(agg aggregateT) IndexT {
	index := make(IndexT, 0, len(agg))
	for _, p := range sortedAggregateKeys(agg) {
		values := sortValues(agg[p], p.valueType)
		index = append(index, IndexEntry{p.key, p.valueType, values})
	}
	return index
}

func printIndex(index IndexT) {
	for _, ref := range index {
		fmt.Printf("%v:", ref.key)
		for _, val := range ref.values {
			fmt.Print(val, " ")
		}
	}
	fmt.Printf("\n")
}

func serializeIndex(index IndexT) []byte {
	stringSep := byte(NUL)

	appendInt := func(buff *bytes.Buffer, i size_t) {
		binary.Write(buff, binary.BigEndian, i)
	}

	appendFloat := func(buff *bytes.Buffer, f float64) {
		binary.Write(buff, binary.BigEndian, f)
	}

	appendFileRefs := func(buff *bytes.Buffer, fileRefs []size_t) {
		appendInt(buff, size_t(len(fileRefs))) // {n file indexes}
		for _, fileRef := range fileRefs {     // {file indexes}
			appendInt(buff, fileRef)
		}
	}
	appendFloatRefs := func(buff *bytes.Buffer, valueRefs []ValueRefs) {
		appendInt(buff, size_t(len(valueRefs))) // {n of values}
		for _, valueRef := range valueRefs {
			appendFloat(buff, valueRef.value.(float64)) // {value}
			appendFileRefs(buff, valueRef.refs)
		}
	}

	appendStringRefs := func(buff *bytes.Buffer, valueRefs []ValueRefs) {
		appendInt(buff, size_t(len(valueRefs))) // {n of values}
		for _, valueRef := range valueRefs {
			buff.WriteString(valueRef.value.(string)) // {value}
			buff.WriteByte(stringSep)                 // {string sep}
			appendFileRefs(buff, valueRef.refs)
		}
	}

	appendNullRefs := func(buff *bytes.Buffer, valueRefs []ValueRefs) {
		for _, valueRef := range valueRefs {
			appendFileRefs(buff, valueRef.refs)
		}
	}

	buff := bytes.NewBuffer(make([]byte, 0, 512))

	for _, indexEntry := range index {
		buff.WriteString(indexEntry.key)           // {key}
		buff.WriteByte(stringSep)                  // {string sep}
		buff.WriteByte(byte(indexEntry.valueType)) // {type byte = 'f' | 's' | 'n'}

		switch indexEntry.valueType {
		case FloatType:
			appendFloatRefs(buff, indexEntry.values)
		case StrType:
			appendStringRefs(buff, indexEntry.values)
		case NullType:
			appendNullRefs(buff, indexEntry.values)
		default:
			message := fmt.Sprintf("Unknown type %c\n", indexEntry.valueType)
			panic(message)
		}
	}
	return buff.Bytes()
}

func deserializeIndex(bytes []byte) IndexT {
	stringSep := byte(NUL)

	readStr := func(bytes []byte, pos int) (string, int) {
		endPos := pos
		for ; endPos < len(bytes); endPos++ {
			if bytes[endPos] == stringSep {
				break
			}
		}

		s := string(bytes[pos:endPos])
		return s, endPos + 1
	}

	readType := func(bytes []byte, pos int) (IndexEntryType, int) {
		entryType := IndexEntryType(bytes[pos])
		return entryType, pos + 1
	}

	readFloat := func(bytes []byte, pos int) (float64, int) {
		endPos := pos + 8
		bits := binary.BigEndian.Uint64(bytes[pos:endPos])
		float := math.Float64frombits(bits)
		return float, endPos
	}

	readInt := func(bytes []byte, pos int) (size_t, int) {
		endPos := pos + 4
		intVal := binary.BigEndian.Uint32(bytes[pos:endPos])
		return size_t(intVal), endPos
	}

	readFileRefs := func(bytes []byte, pos int) ([]size_t, int) {
		nIntRefs, newPos := readInt(bytes, pos)
		pos = newPos
		refs := make([]size_t, 0, nIntRefs)
		for i := 0; i < int(nIntRefs); i++ {
			intRef, newPos := readInt(bytes, pos)
			pos = newPos
			refs = append(refs, intRef)
		}

		return refs, pos
	}

	readFloatValueRefs := func(bytes []byte, pos int, nValues size_t) ([]ValueRefs, int) {
		values := make([]ValueRefs, 0, nValues)

		for i := size_t(0); i < nValues; i++ {
			floatVal, newPos := readFloat(bytes, pos)
			pos = newPos
			refs, newPos := readFileRefs(bytes, pos)
			pos = newPos
			valueRefs := ValueRefs{floatVal, refs}
			values = append(values, valueRefs)
		}

		return values, pos
	}

	readStrValueRefs := func(bytes []byte, pos int, nValues size_t) ([]ValueRefs, int) {
		values := make([]ValueRefs, 0, nValues)

		for i := size_t(0); i < nValues; i++ {
			str, newPos := readStr(bytes, pos)
			pos = newPos
			refs, newPos := readFileRefs(bytes, pos)
			pos = newPos
			valueRefs := ValueRefs{str, refs}
			values = append(values, valueRefs)
		}

		return values, pos
	}

	readNullValueRefs := func(bytes []byte, pos int) ([]ValueRefs, int) {
		refs, pos := readFileRefs(bytes, pos)
		return []ValueRefs{{nil, refs}}, pos
	}

	index := make(IndexT, 0, 32)
	for initPos := 0; initPos < len(bytes); {
		key, pos := readStr(bytes, initPos)
		entryType, pos := readType(bytes, pos)

		switch entryType {
		case FloatType:
			nValues, newPos := readInt(bytes, pos)
			values, newPos := readFloatValueRefs(bytes, newPos, nValues)
			pos = newPos
			entry := IndexEntry{key, entryType, values}
			index = append(index, entry)
		case StrType:
			nValues, newPos := readInt(bytes, pos)
			values, newPos := readStrValueRefs(bytes, newPos, nValues)
			pos = newPos
			entry := IndexEntry{key, entryType, values}
			index = append(index, entry)
		case NullType:
			values, newPos := readNullValueRefs(bytes, pos)
			pos = newPos
			entry := IndexEntry{key, entryType, values}
			index = append(index, entry)
		}

		initPos = pos
	}

	return index
}

func IndexFiles(filePaths []string) IndexT { // todo err
	indexAggregator := make(aggregateT)
	for fileIdx, path := range filePaths {
		unflatten := parseJson(readFile(path))
		flatten := flattenJson(unflatten)
		aggregateJson(indexAggregator, flatten, size_t(fileIdx))
	}

	index := indexAgregate(indexAggregator)
	return index
}

func SaveIndex(index IndexT, dirPath string) error {
	indexBytes := serializeIndex(index)
	err := os.WriteFile(dirPath+"/INDEX", indexBytes, 0644)
	return err
}

func ReadIndex(dirPath string) IndexT { // TODO err
	indexBytes := readFile(dirPath + "/INDEX")
	index := deserializeIndex(indexBytes)
	return index
}

type valueTypes interface{ float64 | string }

type valueQuery[T valueTypes] struct {
	Key   string
	Type  IndexEntryType
	Value T
}

type nullQuery struct {
	Key string
}

func queryForNullRefs(index *IndexT, query *nullQuery) []size_t {
	indexEntryCmp := func(entry IndexEntry, key string) int {
		return valueWithTypeCmp(entry.key, key, entry.valueType, NullType)
	}

	if entryIdx, found := slices.BinarySearchFunc(*index, query.Key, indexEntryCmp); found {
		entry := (*index)[entryIdx]
		return entry.values[0].refs
	}
	return nil
}

func getFileRefs[T valueTypes](index *IndexT, query *valueQuery[T]) fileRefs {
	// TODO propagate fileRefs = bitflags.BitFlags Type for file indexes throughout the project
	refsArrTofileRefs := func(refs []size_t) fileRefs {
		fr := fileRefs{}
		for _, ref := range refs {
			fr.Set(uint(ref))
		}
		return fr
	}

	indexEntryCmp := func(entry IndexEntry, key string) int {
		return valueWithTypeCmp(entry.key, key, entry.valueType, query.Type)
	}

	valueRefCmp := func(valueRef ValueRefs, key T) int {
		return cmp.Compare(valueRef.value.(T), key)
	}

	if entryIdx, found := slices.BinarySearchFunc(*index, query.Key, indexEntryCmp); found {
		entry := (*index)[entryIdx]
		if refIdx, found := slices.BinarySearchFunc(entry.values, query.Value, valueRefCmp); found {
			return refsArrTofileRefs(entry.values[refIdx].refs)
		}
	}
	return refsArrTofileRefs(nil)
}

// func refsUnion(refs ...[]size_t) []size_t {
// 	/* Result of AND operations on refs */
// 	refsMask := bitflags.BitsBlockEmpty
// 	for _, refsArr := range refs {
// 		for _, ref := range refsArr {
// 			refsMask = refsMask.Union(bitflags.BitsBlockEmpty.Set(uint(ref)))
// 		}
// 	}

// 	res := make([]size_t, 0, bitflags.BitsBlockSize)

// 	for i := 0; i < bitflags.BitsBlockSize; i++ {
// 		if refsMask.Has(uint(i)) {
// 			res = append(res, size_t(i))
// 		}
// 	}
// 	return res
// }

// func refsIntersection(refs ...[]size_t) []size_t {
// 	/* Result of OR operations on refs */
// 	refsMask := bitflags.BitsBlockFull
// 	for _, refsArr := range refs {
// 		union := bitflags.BitsBlockEmpty
// 		for _, ref := range refsArr {
// 			union = union.Union(bitflags.BitsBlockEmpty.Set(uint(ref)))
// 		}
// 		refsMask = refsMask & union
// 	}
// 	// fmt.Printf("%b\n", refsMask)
// 	res := make([]size_t, 0, bitflags.BitsBlockSize)

// 	for i := 0; i < bitflags.BitsBlockSize; i++ {
// 		if refsMask.Has(uint(i)) {
// 			res = append(res, size_t(i))
// 		}
// 	}

// 	return res
// }

// stack based refs operations: unions and intersections
// *******>>
type fileRefs = bitflags.BitFlags
type refStack struct {
	stack []fileRefs
}

func (s *refStack) Push(f fileRefs) {
	s.stack = append(s.stack, f)
}
func (s *refStack) Pop() fileRefs {
	l := len(s.stack) - 1
	res := s.stack[l]
	s.stack = s.stack[:l]
	return res
}
func (s *refStack) And(f fileRefs) {
	l := len(s.stack) - 1
	s.stack[l] = s.stack[l].Intersect(f)
}
func (s *refStack) Or(f fileRefs) {
	l := len(s.stack) - 1
	s.stack[l] = s.stack[l].Union(f)
}

// <<*******

func QueryIndex(index *IndexT, query string) {
	// fmt.Println(getFileRefs(index, &valueQuery[string]{"/social/twitter", StrType, "https://twitter.com"}))
	// fmt.Println(getFileRefs(index, &valueQuery[float64]{"/age", FloatType, 23}))

	// fmt.Println(queryForNullRefs(index, &nullQuery{"/now null behaves"}))

	// for i, k := range queryForNullRefs(index, &nullQuery{"/not found"}) {
	// 	fmt.Println(i, k)
	// 	fmt.Println(queryForNullRefs(index, &nullQuery{"/not found"}))
	// }

	refsToSlice := func(refs fileRefs) []uint {
		refsSlice := make([]uint, 0, 32)
		for av := range refs.Traverse() {
			refsSlice = append(refsSlice, av)
		}
		return refsSlice
	}

	/* Example query SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com' */
	currStack := refStack{}
	currStack.Push(getFileRefs(index, &valueQuery[string]{"/social/twitter", StrType, "https://twitter.com"}))
	fmt.Printf("File refs for\nSELECT * FROM c WHERE c.social.twitter = 'https://twitter.com'\n%v\n", refsToSlice(currStack.Pop()))

	/* Example query SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com' AND c.social.facebook = 'https://facebook.com' */
	currStack = refStack{}
	currStack.Push(getFileRefs(index, &valueQuery[string]{"/social/twitter", StrType, "https://twitter.com"}))
	currStack.And(getFileRefs(index, &valueQuery[string]{"/social/facebook", StrType, "https://facebook.com"}))
	fmt.Printf("File refs for\nSELECT * FROM c WHERE c.social.twitter = 'https://twitter.com' AND c.social.facebook = 'https://facebook.com'\n%v\n", refsToSlice(currStack.Pop()))

	/* Example query SELECT * FROM c WHERE c.age = 23 OR c.age = 17 */
	currStack = refStack{}
	currStack.Push(getFileRefs(index, &valueQuery[float64]{"/age", FloatType, 23}))
	currStack.Or(getFileRefs(index, &valueQuery[float64]{"/age", FloatType, 17}))
	fmt.Printf("File refs for\nSELECT * FROM c WHERE c.age = 23 OR c.age = 17\n%v\n", refsToSlice(currStack.Pop()))

}

func main() {
	// paths := []string{"./db/0", "./db/1"} // todo use listDir(dirPath)
	// index := IndexFiles(paths)

	// err := SaveIndex(index, "./db")
	// check(err)

	index := ReadIndex("./db")
	QueryIndex(&index, "")
}
