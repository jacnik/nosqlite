package main

import (
	"testing"
)

func compareValueRefs(expected, actual ValueRefs, valueType IndexEntryType) bool {
	switch valueType {
	case FloatType:
		if expected.value.(float64) != actual.value.(float64) {
			return false
		}
	case StrType:
		if expected.value.(string) != actual.value.(string) {
			return false
		}
	case NullType:
		if expected.value != nil || actual.value != nil {
			return false
		}
	default:
		return false
	}
	if len(expected.refs) != len(actual.refs) {
		return false
	}
	for i, v := range expected.refs {
		if v != actual.refs[i] {
			return false
		}
	}

	return true
}

func compareIndexEntries(expected, actual IndexEntry) bool {
	if expected.valueType != actual.valueType {
		return false
	}
	if expected.key != actual.key {
		return false
	}
	if len(expected.values) != len(actual.values) {
		return false
	}
	for i, v := range expected.values {
		if !compareValueRefs(v, actual.values[i], actual.valueType) {
			return false
		}
	}
	return true
}

func compareIndexes(expected, actual IndexT) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i, v := range expected {
		if !compareIndexEntries(v, actual[i]) {
			return false
		}
	}
	return true
}

// Check if index can be serialized and deseralized back.
func TestSerializeAndDeserializeIndex(t *testing.T) {
	paths := []string{"./db/0", "./db/1"}
	index := IndexFiles(paths)

	indexBytes := serializeIndex(index)

	deserializedIndex := deserializeIndex(indexBytes)

	if !compareIndexes(index, deserializedIndex) {
		t.Fatalf("Deserialized index different than original:\n%v\n%v", index, deserializedIndex)
	}
}

// Check if it can create correct index from files.
func TestIndexFiles(t *testing.T) {
	paths := []string{"./db/0", "./db/1"}
	index := IndexFiles(paths)

	expected := IndexT{
		IndexEntry{"/age", FloatType, []ValueRefs{{value: 17.0, refs: []size_t{1}}, {value: 23.0, refs: []size_t{0}}}},
		IndexEntry{"/arr/0", FloatType, []ValueRefs{{value: 2.0, refs: []size_t{0}}}},
		IndexEntry{"/arr/1", FloatType, []ValueRefs{{value: 3.0, refs: []size_t{0}}}},
		IndexEntry{"/name", StrType, []ValueRefs{{value: "Elliot", refs: []size_t{0}}, {value: "Fraser", refs: []size_t{1}}}},
		IndexEntry{"/now null behaves", NullType, []ValueRefs{{value: nil, refs: []size_t{0}}}},
		IndexEntry{"/social/facebook", StrType, []ValueRefs{{value: "https://facebook.com", refs: []size_t{0, 1}}}},
		IndexEntry{"/social/twitter", StrType, []ValueRefs{{value: "https://twitter.com", refs: []size_t{0, 1}}}},
		IndexEntry{"/type", StrType, []ValueRefs{{value: "Author", refs: []size_t{1}}, {value: "Reader", refs: []size_t{0}}}}}

	if !compareIndexes(index, expected) {
		t.Fatalf("Expected index different than actual:\n%v\n%v", index, expected)
	}
}
