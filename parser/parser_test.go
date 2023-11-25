package parser

import (
	"testing"
)

func compareTokens(actual, expected []token) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i, a := range actual {
		if expected[i] != a {
			return false
		}
	}
	return true
}

// Tokenizer: Check if simple query will be tokenized correctly.
func TestTokenizeSimple(t *testing.T) {
	query := "SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com'"

	tokens, _ := tokenize(query)
	expected := []token{
		{select_, nil},
		{star, nil},
		{from, nil},
		{ident, "c"},
		{where, nil},
		{ident, "c"},
		{dot, nil},
		{ident, "social"},
		{dot, nil},
		{ident, "twitter"},
		{eq, nil},
		{text, "https://twitter.com"},
		{eof, nil},
	}

	if !compareTokens(tokens, expected) {
		t.Fatalf("Got tokens different than expected:\n%v\n%v", tokens, expected)
	}
}

// Tokenizer: Check if simple query with AND clause will be tokenized correctly.
func TestTokenizeSimpleQueryWithAnd(t *testing.T) {
	query := "SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com' AND c.age > 17.0"

	tokens, _ := tokenize(query)
	expected := []token{
		{select_, nil},
		{star, nil},
		{from, nil},
		{ident, "c"},
		{where, nil},
		{ident, "c"},
		{dot, nil},
		{ident, "social"},
		{dot, nil},
		{ident, "twitter"},
		{eq, nil},
		{text, "https://twitter.com"},
		{and, nil},
		{ident, "c"},
		{dot, nil},
		{ident, "age"},
		{gt, nil},
		{float, 17.0},
		{eof, nil},
	}

	if !compareTokens(tokens, expected) {
		t.Fatalf("Got tokens different than expected:\n%v\n%v", tokens, expected)
	}
}

// Tokenizer: Check if simple query with OR clause will be tokenized correctly.
func TestTokenizeSimpleQueryWithOr(t *testing.T) {
	query := "SELECT * FROM c WHERE c.age = 23 OR c.age = 17"

	tokens, _ := tokenize(query)
	expected := []token{
		{select_, nil},
		{star, nil},
		{from, nil},
		{ident, "c"},
		{where, nil},
		{ident, "c"},
		{dot, nil},
		{ident, "age"},
		{eq, nil},
		{float, float64(23)},
		{or, nil},
		{ident, "c"},
		{dot, nil},
		{ident, "age"},
		{eq, nil},
		{float, float64(17)},
		{eof, nil},
	}

	if !compareTokens(tokens, expected) {
		t.Fatalf("Got tokens different than expected:\n%v\n%v", tokens, expected)
	}
}

func comparePrograms(actual, expected Program) bool {
	if len(actual.Instructions) != len(expected.Instructions) {
		return false
	}
	for i, a := range actual.Instructions {
		if expected.Instructions[i] != a {
			return false
		}
	}
	return true
}

// Parse: Check if simple query will be parsed correctly.
func TestParseSimple(t *testing.T) {
	query := "SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com'"

	program, _ := Parse(query)
	expected := Program{[]Instruction{
		{Push, "/social/twitter", Eq, "https://twitter.com"},
	}}

	if !comparePrograms(program, expected) {
		t.Fatalf("Got programs different than expected:\n%v\n%v", program, expected)
	}
}

// Parse: Check if simple query with OR clause will be parsed correctly.
func TestParseSimpleWithOr(t *testing.T) {
	query := "SELECT * FROM c WHERE c.age = 23 OR c.age = 17"

	program, _ := Parse(query)
	expected := Program{Instructions: []Instruction{
		{Push, "/age", Eq, float64(23)},
		{Or, "/age", Eq, float64(17)},
	}}

	if !comparePrograms(program, expected) {
		t.Fatalf("Got programs different than expected:\n%v\n%v", program, expected)
	}
}

// Parse: Check if simple query with AND clause will be parsed correctly.
func TestParseSimpleWithAnd(t *testing.T) {
	query := "SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com' AND c.social.facebook = 'https://facebook.com'"

	program, _ := Parse(query)
	expected := Program{Instructions: []Instruction{
		{Push, "/social/twitter", Eq, "https://twitter.com"},
		{And, "/social/facebook", Eq, "https://facebook.com"},
	}}

	if !comparePrograms(program, expected) {
		t.Fatalf("Got programs different than expected:\n%v\n%v", program, expected)
	}
}
