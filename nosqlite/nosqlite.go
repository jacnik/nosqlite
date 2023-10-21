package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
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

func main() {
	dirPath := "./db"

	agg := make(map[string]map[interface{}][]int)
	for fileIdx, fileName := range listDir(dirPath) {
		fmt.Println("Adding file:", fileName)

		unflatten := parseJson(readFile(dirPath + "/" + strconv.Itoa(fileIdx)))
		flatten := flattenJson(unflatten)
		aggregateJson(agg, flatten, fileIdx)
	}

	fmt.Println(agg)
}
