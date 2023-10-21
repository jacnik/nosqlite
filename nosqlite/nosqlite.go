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
		if !entry.IsDir() {
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

func flattenJsonMap(acc map[string]interface{}, prefix string, jMap map[string]interface{}) {
	for key, jItem := range jMap {
		flattenWithPrefix(acc, prefix+"/"+key, jItem)
	}
}

func flattenJsonArr(acc map[string]interface{}, prefix string, jArr []interface{}) {
	for i, jItem := range jArr {
		flattenWithPrefix(acc, prefix+"/"+strconv.Itoa(i), jItem)
	}
}

func flattenWithPrefix(acc map[string]interface{}, prefix string, unflatten interface{}) {
	switch v := unflatten.(type) {
	case map[string]interface{}:
		flattenJsonMap(acc, prefix, v)
	case []interface{}:
		flattenJsonArr(acc, prefix, v)
	default:
		acc[prefix] = v
	}
}

func flattenJson(unflatten interface{}) map[string]interface{} {
	flatten := make(map[string]interface{})
	flattenWithPrefix(flatten, "", unflatten)
	return flatten
}

func aggregateJson(agg map[string]map[interface{}][]int, flatten map[string]interface{}, fileIdx int) {
	for prop, value := range flatten {
		propMap, hasPropMap := agg[prop]
		if !hasPropMap {
			propMap = make(map[interface{}][]int)
			propMap[value] = []int{fileIdx}
			agg[prop] = propMap
		} else {
			propMap[value] = append(propMap[value], fileIdx)
		}
	}
}

type valueRefs struct {
	value interface{}
	refs  []int
}

type propRefs struct {
	prop   string
	values []valueRefs
}

func sortValues(values map[interface{}][]int) []valueRefs {
	lessStr := func(a string, b interface{}) int {
		switch v := b.(type) {
		case string:
			return cmp.Compare(a, v)
		case float64:
			return 1 // floats compared to strings will be treated as smaller
		}
		return 1 // b is most likely nil
	}

	lessFloat := func(a float64, b interface{}) int {
		switch v := b.(type) {
		case float64:
			return cmp.Compare(a, v)
		case string:
			return -1 // floats compared to strings will be treated as smaller
		}
		return 1 // b is most likely nil
	}

	less := func(a, b interface{}) int {
		/* Should return a negative number when a < b,
		** a positive number when a > b
		** and zero when a == b. */
		switch v := a.(type) {
		case string:
			return lessStr(v, b)
		case float64:
			return lessFloat(v, b)
		}
		return -1 // a is most likely nil
	}

	keys := make([]interface{}, 0, len(values))
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

func main() {
	dirPath := "./db"

	agg := make(map[string]map[interface{}][]int)
	for fileIdx, fileName := range listDir(dirPath) {
		fmt.Println("Adding file:", fileName)

		unflatten := parseJson(readFile(dirPath + "/" + strconv.Itoa(fileIdx)))
		flatten := flattenJson(unflatten)
		aggregateJson(agg, flatten, fileIdx)
	}

	// todo sort props fn
	properties := make([]string, 0, len(agg))
	for p := range agg {
		properties = append(properties, p)
	}
	slices.SortFunc(properties, cmp.Compare)

	index := make([]propRefs, 0, len(properties))
	for _, p := range properties {
		values := sortValues(agg[p])
		index = append(index, propRefs{p, values})
	}

	fmt.Println(index)
}
