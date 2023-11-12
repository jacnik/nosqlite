package parser

import (
	"strconv"
)

type OpType byte

const (
	Eq OpType = '='
	Gt OpType = '>'
	Lt OpType = '<'
)

type InstructionType byte

const (
	Push InstructionType = 'p'
	And  InstructionType = 'a'
	Or   InstructionType = 'o'
)

type Instruction struct {
	Type     InstructionType
	QueryKey string
	Op       OpType
	QueryVal interface{}
}

type Program struct {
	Instructions []Instruction
}

// fmt.Println(queryForNullRefs(index, &nullQuery{"/now null behaves"}))

// for i, k := range queryForNullRefs(index, &nullQuery{"/not found"}) {
// 	fmt.Println(i, k)
// 	fmt.Println(queryForNullRefs(index, &nullQuery{"/not found"}))
// }

func Parse(query string) (Program, error) {
	/* SELECT * FROM c WHERE c.age = 23 OR c.age = 17 */
	program := Program{Instructions: []Instruction{
		{Push, "/age", Eq, 23},
		{Or, "/age", Eq, 17},
	}}

	/* Example query SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com' */
	program = Program{Instructions: []Instruction{
		{Push, "/social/twitter", Eq, "https://twitter.com"},
	}}

	/* SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com' AND c.social.facebook = 'https://facebook.com' */
	program = Program{Instructions: []Instruction{
		{Push, "/social/twitter", Eq, "https://twitter.com"},
		{And, "/social/facebook", Eq, "https://facebook.com"},
	}}

	return program, nil
}

type tokenType byte

const (
	text tokenType = iota
	integer
	float
	ident
	sel // select
	from
	where
	comma
	//	operator
	star
	eof
	lparem
	rparem
	lsqbrack
	rsqbrack
	dot
	eq
	gt
	lt
	and
	or
)

type token struct {
	tokenType tokenType
	value     any
}

func tokenize(query string) []token {
	isWhitespace := func(query string, i int) bool {
		spaces := []byte{' ', '\t', '\n'}
		for _, s := range spaces {
			if query[i] == s {
				return true
			}
		}
		return false
	}

	isIdentChar := func(query string, i int) (ok bool) {
		c := query[i]
		if c >= 'a' && c <= 'z' {
			return true
		}
		if c >= 'A' && c <= 'Z' {
			return true
		}
		if c == '_' {
			return true
		}
		return false
	}

	classifyIdent := func(identifier string) token {

		switch identifier {
		case "SELECT":
			return token{sel, nil}
		case "FROM":
			return token{from, nil}
		case "WHERE":
			return token{where, nil}
		case "AND":
			return token{and, nil}
		case "OR":
			return token{or, nil}

		}

		return token{ident, identifier}
	}

	appendIdent := func(tokens *[]token, query string, i int) (newPos int) {
		j := i
		for ; j < len(query) && isIdentChar(query, j); j++ {
		}
		if i != j {
			*tokens = append(*tokens, classifyIdent(query[i:j]))
		}

		return j
	}

	appendSpecial := func(tokens *[]token, query string, i int) (newPos int) {
		switch query[i] {
		case '*':
			*tokens = append(*tokens, token{star, nil})
			return i + 1
		case '.':
			*tokens = append(*tokens, token{dot, nil})
			return i + 1
		case '=':
			*tokens = append(*tokens, token{eq, nil})
			return i + 1
		case '>':
			*tokens = append(*tokens, token{gt, nil})
			return i + 1
		case '<':
			*tokens = append(*tokens, token{lt, nil})
			return i + 1
		}
		return i
	}

	appendText := func(tokens *[]token, query string, i int) (newPos int) {
		if query[i] != '\'' {
			return i
		}
		for j := i + 1; j < len(query); j++ {
			if query[j] == '\'' {
				*tokens = append(*tokens, token{text, query[i+1 : j]})
				return j + 1
			}
		}
		return len(query)
	}

	appendNumber := func(tokens *[]token, query string, i int) (newPos int) {
		// TODO floats starting with `.`, ints and negative numbers
		isDigit := func(query string, i int) bool {
			return query[i] >= '0' && query[i] <= '9'
		}
		if !isDigit(query, i) {
			return i
		}

		dotsCount := 0
		for j := i; j < len(query); j++ {
			if query[j] == '.' {
				dotsCount++
				j++
			}
			if dotsCount > 1 {
				panic("Invalid syntax. Too many dots in number") // TODO error
			}
			if !isDigit(query, j) {
				if i == j {
					return i
				} else if f, err := strconv.ParseFloat(query[i:j], 64); err == nil {
					*tokens = append(*tokens, token{float, f})
					return j + 1
				}
				panic("Unable to cast to float")
			}
			if j >= len(query)-1 {
				if f, err := strconv.ParseFloat(query[i:j], 64); err == nil {
					*tokens = append(*tokens, token{float, f})
					return j + 1
				}
			}
		}

		return i
	}

	tokens := make([]token, 0, 8)
	for i := 0; ; {
		if isWhitespace(query, i) {
			i++
		}

		i = appendIdent(&tokens, query, i)
		i = appendSpecial(&tokens, query, i)
		i = appendText(&tokens, query, i)
		i = appendNumber(&tokens, query, i)

		if i >= len(query) {
			tokens = append(tokens, token{eof, nil})
			break
		}
	}

	return tokens
}
