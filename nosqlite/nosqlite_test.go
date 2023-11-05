package main

import (
	"testing"

	"github.com/jacnik/nosqlite/parser"
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

func compareSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

func compareRefs(a, b fileRefs) bool {
	aslice := make([]uint, 0, 32)
	for av := range a.Traverse() {
		aslice = append(aslice, av)
	}

	bslice := make([]uint, 0, 32)
	for bv := range b.Traverse() {
		bslice = append(bslice, bv)
	}
	return compareSlices(aslice, bslice)
}

// Check if it can return file idx list for simple string query.
func TestQueryIndexForString(t *testing.T) {
	index := ReadIndex("./db")
	refs := getFileRefs(&index, "/social/twitter", parser.Eq, "https://twitter.com", StrType)

	expected := fileRefs{}
	expected.Set(0, 1)

	if !compareRefs(refs, expected) {
		t.Fatalf("Expected refs different than actual:\n%v\n%v", expected, refs)
	}
}

// Check if it can return file idx list for simple float query.
func TestQueryIndexForFloat(t *testing.T) {
	index := ReadIndex("./db")
	refs := getFileRefs(&index, "/age", parser.Eq, 23, FloatType)

	expected := fileRefs{}
	expected.Set(0)

	if !compareRefs(refs, expected) {
		t.Fatalf("Expected refs different than actual:\n%v\n%v", expected, refs)
	}
}

// Check if it can return file idx list for simple null query.
func TestQueryIndexForNull(t *testing.T) {
	index := ReadIndex("./db")
	refs := queryForNullRefs(&index, &nullQuery{"/now null behaves"})

	expected := []size_t{0}

	if !compareSlices(refs, expected) {
		t.Fatalf("Expected refs different than actual:\n%v\n%v", expected, refs)
	}
}

// Check if it can return empty list when querying for non existing element.
func TestQueryIndexForNonExisting(t *testing.T) {
	index := ReadIndex("./db")
	refs := queryForNullRefs(&index, &nullQuery{"/not found"})

	expected := []size_t{}

	if !compareSlices(refs, expected) {
		t.Fatalf("Expected refs different than actual:\n%v\n%v", expected, refs)
	}
}
