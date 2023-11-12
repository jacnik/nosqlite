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

// Check if simple query will be tokenized correctly.
func TestTokenizeSimpleQuery(t *testing.T) {
	query := "SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com' AND c.age > 17.0"

	tokens := tokenize(query)
	expected := []token{
		{sel, nil},
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
