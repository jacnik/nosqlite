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

type indexType byte

const (
	floatType indexType = 'f'
	strType   indexType = 's'
	nullType  indexType = 'n'
)

type valueRefs struct {
	value jsonTypeValue
	refs  []int
}

func (v valueRefs) String() string {
	return fmt.Sprintf("%v %v", v.value, v.refs)
}

type propRefs struct {
	prop   string
	values []valueRefs
}

type jsonTypeValue struct {
	_type indexType
	value interface{}
}

func (p jsonTypeValue) String() string {
	return fmt.Sprintf("%v{%c}", p.value, p._type)
}

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

func flattenJsonMap(acc map[string]jsonTypeValue, prefix string, jMap map[string]interface{}) {
	for key, jItem := range jMap {
		flattenWithPrefix(acc, prefix+"/"+key, jItem)
	}
}

func flattenJsonArr(acc map[string]jsonTypeValue, prefix string, jArr []interface{}) {
	for i, jItem := range jArr {
		flattenWithPrefix(acc, prefix+"/"+strconv.Itoa(i), jItem)
	}
}

func flattenWithPrefix(acc map[string]jsonTypeValue, prefix string, unflatten interface{}) {
	switch v := unflatten.(type) {
	case map[string]interface{}:
		flattenJsonMap(acc, prefix, v)
	case []interface{}:
		flattenJsonArr(acc, prefix, v)
	case string:
		acc[prefix] = jsonTypeValue{strType, v}
	case float64:
		acc[prefix] = jsonTypeValue{floatType, v}
	default:
		acc[prefix] = jsonTypeValue{nullType, v}
	}
}

func flattenJson(unflatten interface{}) map[string]jsonTypeValue {
	flatten := make(map[string]jsonTypeValue)
	flattenWithPrefix(flatten, "", unflatten)
	return flatten
}

func aggregateJson(agg map[string]map[jsonTypeValue][]int, flatten map[string]jsonTypeValue, fileIdx int) {
	for prop, value := range flatten {
		propMap, hasPropMap := agg[prop]
		if !hasPropMap {
			propMap = make(map[jsonTypeValue][]int)
			propMap[value] = []int{fileIdx}
			agg[prop] = propMap
		} else {
			propMap[value] = append(propMap[value], fileIdx)
		}
	}
}

func sortValues(values map[jsonTypeValue][]int) []valueRefs {
	less := func(a, b jsonTypeValue) int {
		/* Should return a negative number when a < b,
		** a positive number when a > b
		** and zero when a == b. */
		typesCmp := cmp.Compare(a._type, b._type)
		if typesCmp != 0 {
			return typesCmp
		}

		switch a._type {
		case floatType:
			return cmp.Compare(a.value.(float64), b.value.(float64))
		case strType:
			return cmp.Compare(a.value.(string), b.value.(string))
		case nullType:
			return 0
		}
		message := fmt.Sprintf("Unknown type %c", a._type)
		panic(message)
	}

	keys := make([]jsonTypeValue, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}

	slices.SortFunc(keys, less)

	sortedRefs := make([]valueRefs, 0, len(keys))
	for _, key := range keys {
		sortedRefs = append(sortedRefs, valueRefs{key, values[key]})
	}
	return sortedRefs
}

func sortedKeys[K cmp.Ordered, V any](mapToSort map[K]V) []K {
	keys := make([]K, 0, len(mapToSort))
	for p := range mapToSort {
		keys = append(keys, p)
	}
	slices.SortFunc(keys, cmp.Compare)
	return keys
}

func createIndex(referenceAgg map[string]map[jsonTypeValue][]int) []propRefs {
	index := make([]propRefs, 0, len(referenceAgg))
	for _, p := range sortedKeys(referenceAgg) {
		values := sortValues(referenceAgg[p])
		index = append(index, propRefs{p, values})
	}
	return index
}

func printIndex(index []propRefs) {
	// fmt.Println(index)

	for _, ref := range index {
		fmt.Printf("%v:", ref.prop)
		for _, val := range ref.values {
			fmt.Print(val, " ")
		}
	}
	fmt.Printf("\n")
}

func main() {
	dirPath := "./db"

	agg := make(map[string]map[jsonTypeValue][]int)
	for fileIdx, fileName := range listDir(dirPath) {
		fmt.Println("Adding file:", fileName)

		unflatten := parseJson(readFile(dirPath + "/" + strconv.Itoa(fileIdx)))
		flatten := flattenJson(unflatten)
		aggregateJson(agg, flatten, fileIdx)
	}

	index := createIndex(agg)
	printIndex(index)
}
