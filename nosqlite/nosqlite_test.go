package main

import (
	"strconv"
	"testing"
)

func compareValueRefs(v1, v2 ValueRefs, valueType IndexEntryType) bool {
	switch valueType {
	case FloatType:
		if v1.value.(float64) != v2.value.(float64) {
			return false
		}
	case StrType:
		if v1.value.(string) != v2.value.(string) {
			return false
		}
	case NullType:
		if v1.value != nil || v2.value != nil {
			return false
		}
	default:
		return false
	}
	if len(v1.refs) != len(v2.refs) {
		return false
	}
	for i, v := range v1.refs {
		if v != v2.refs[i] {
			return false
		}
	}

	return true
}

func compareIndexEntries(e1, e2 IndexEntry) bool {
	if e1.valueType != e2.valueType {
		return false
	}
	if e1.key != e2.key {
		return false
	}
	if len(e1.values) != len(e2.values) {
		return false
	}
	for i, v := range e1.values {
		if !compareValueRefs(v, e2.values[i], e2.valueType) {
			return false
		}
	}
	return true
}

func compareIndexes(i1, i2 IndexT) bool {
	if len(i1) != len(i2) {
		return false
	}
	for i, v := range i1 {
		if !compareIndexEntries(v, i2[i]) {
			return false
		}
	}
	return true
}

// Check if index can be serialized and deseralized back.
func TestSerializeAndDeserializeIndex(t *testing.T) {
	dirPath := "./db"

	indexAggregator := make(aggregateT)
	for fileIdx, _ := range listDir(dirPath) {

		unflatten := parseJson(readFile(dirPath + "/" + strconv.Itoa(fileIdx)))
		flatten := flattenJson(unflatten)
		aggregateJson(indexAggregator, flatten, int32(fileIdx))
	}

	index := createIndex(indexAggregator)

	indexBytes := serializeIndex(index)

	deserializedIndex := deserializeIndex(indexBytes)

	if !compareIndexes(index, deserializedIndex) {
		t.Fatalf("Deserialized index different than original:\n%v\n%v", index, deserializedIndex)
	}
}
