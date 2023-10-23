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

// func sortedKeys[K cmp.Ordered, V any](mapToSort map[K]V) []K {
// 	keys := make([]K, 0, len(mapToSort))
// 	for p := range mapToSort {
// 		keys = append(keys, p)
// 	}
// 	slices.SortFunc(keys, cmp.Compare)
// 	return keys
// }

func sortedAggregateKeys(mapToSort aggregateT) []aggregateKeyT {
	aggregateKeyCmp := func(x, y aggregateKeyT) int {
		valueCmp := cmp.Compare(x.key, y.key)
		if valueCmp != 0 {
			return valueCmp
		}
		return cmp.Compare(x.valueType, y.valueType)
	}
	keys := make([]aggregateKeyT, 0, len(mapToSort))
	for p := range mapToSort {
		keys = append(keys, p)
	}

	slices.SortFunc(keys, aggregateKeyCmp)
	return keys
}

func createIndex(agg aggregateT) IndexT {
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
		// fmt.Printf("Found string %s on positions %d : %d.\n", s, pos, endPos)
		return s, endPos + 1
	}

	readType := func(bytes []byte, pos int) (IndexEntryType, int) {
		entryType := IndexEntryType(bytes[pos])
		// fmt.Printf("Found type %c on position %d : %d.\n", entryType, pos, pos+1)
		return entryType, pos + 1
	}

	readFloat := func(bytes []byte, pos int) (float64, int) {
		endPos := pos + 8
		bits := binary.BigEndian.Uint64(bytes[pos:endPos])
		float := math.Float64frombits(bits)

		// fmt.Printf("Found float %f on position %d : %d\n", float, pos, endPos)
		return float, endPos
	}

	readInt := func(bytes []byte, pos int) (size_t, int) {
		endPos := pos + 4
		intVal := binary.BigEndian.Uint32(bytes[pos:endPos])

		// fmt.Printf("Found int %d on position %d : %d\n", intVal, pos, endPos)
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

func main() {
	dirPath := "./db"

	indexAggregator := make(aggregateT)
	for fileIdx, fileName := range listDir(dirPath) {
		fmt.Println("Adding file:", fileName)

		unflatten := parseJson(readFile(dirPath + "/" + strconv.Itoa(fileIdx)))
		flatten := flattenJson(unflatten)
		aggregateJson(indexAggregator, flatten, size_t(fileIdx))
	}

	index := createIndex(indexAggregator)
	printIndex(index)

	indexBytes := serializeIndex(index)
	// fmt.Println(indexBytes)
	// fmt.Println(len(indexBytes))

	deserializedIndex := deserializeIndex(indexBytes)
	printIndex(deserializedIndex)

	// fmt.Println("Are indexes equal:", index == deserializedIndex)
}
