package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
)

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
	refs  []int32
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
type aggregateFileRefT []int32
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

func aggregateJson(agg aggregateT, flatten flattenJsonT, fileIdx int32) {
	for aggregateKey, aggregateValue := range flatten {
		fileRefsMap, hasFileRefsMap := agg[aggregateKey]
		if !hasFileRefsMap {
			fileRefsMap = make(map[aggregateValueT]aggregateFileRefT)
			fileRefsMap[aggregateValue] = []int32{fileIdx}
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
	// fmt.Println(index)
	for _, ref := range index {
		fmt.Printf("%v:", ref.key)
		for _, val := range ref.values {
			fmt.Print(val, " ")
		}
	}
	fmt.Printf("\n")
}

// func serializeIndex(index []indexEntry) []byte {

// 	appendFloat := func(buff *bytes.Buffer, f float64) {
// 		binary.Write(buff, binary.BigEndian, f)
// 	}
// 	appendInt := func(buff *bytes.Buffer, i int32) {
// 		binary.Write(buff, binary.BigEndian, i)
// 	}

// 	buff := bytes.NewBuffer([]byte{})

// 	stringSep := byte(NUL)
// 	lineEnd := byte(LF)
// 	// maybe buff.Grow(n) .. if you hit perf issues?
// 	for _, prop := range index {
// 		buff.WriteString(prop.prop) // {key}
// 		buff.WriteByte(stringSep)   // \x00

// 		for _, propValue := range prop.values {
// 			switch propValue.value._type {
// 			case floatType:
// 				buff.WriteByte(byte(floatType))                    // {type byte = 'f'}
// 				appendFloat(buff, propValue.value.value.(float64)) // {value}
// 				appendInt(buff, int32(len(propValue.refs)))        // {n file indexes}
// 				for _, fileRef := range propValue.refs {           // {file indexes}
// 					appendInt(buff, int32(fileRef))
// 				}

// 			case strType:
// 				fmt.Printf("Appending strings\n")
// 			case nullType:
// 				fmt.Printf("Appending nulls\n")
// 			default:
// 				message := fmt.Sprintf("Unknown type %c\n", propValue.value._type)
// 				panic(message)
// 			}
// 		}
// 		buff.WriteByte(lineEnd) // {\n}
// 	}
// 	return buff.Bytes()
// }

func main() {
	dirPath := "./db"

	indexAggregator := make(aggregateT)
	for fileIdx, fileName := range listDir(dirPath) {
		fmt.Println("Adding file:", fileName)

		unflatten := parseJson(readFile(dirPath + "/" + strconv.Itoa(fileIdx)))
		flatten := flattenJson(unflatten)
		aggregateJson(indexAggregator, flatten, int32(fileIdx))
	}

	index := createIndex(indexAggregator)
	printIndex(index)
}
