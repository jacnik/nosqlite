package parser

import (
	"strconv"
)

type tokenKind byte

const (
	text tokenKind = iota
	integer
	float
	ident
	select_
	from
	where
	comma
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
	kind  tokenKind
	value any
}

func tokenize(query string) ([]token, error) {
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
			return token{select_, nil}
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

		if i >= len(query)-1 || !isDigit(query, i) {
			return i
		}

		dotsCount := 0
		for j := i + 1; j < len(query); j++ {
			if query[j] == '.' {
				dotsCount++
				j++
			}
			if dotsCount > 1 {
				panic("Invalid syntax. Too many dots in number") // TODO error
			}
			if !isDigit(query, j) {
				if f, err := strconv.ParseFloat(query[i:j], 64); err == nil {
					*tokens = append(*tokens, token{float, f})
					return j + 1
				}
				panic("Unable to cast to float")
			}
			if j >= len(query)-1 {
				if f, err := strconv.ParseFloat(query[i:j+1], 64); err == nil {
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

	return tokens, nil
}

type OpType byte

const (
	Eq OpType = '='
	Gt OpType = '>'
	Lt OpType = '<'
)

type InstructionKind byte

const (
	Push InstructionKind = 'p'
	And  InstructionKind = 'a'
	Or   InstructionKind = 'o'
)

type Instruction struct {
	Kind InstructionKind
	Key  string
	Op   OpType
	Val  interface{}
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
	readSelection := func(tokens []token) (int, int) {
		// TODO implement selection
		if tokens[0].kind != select_ {
			return 0, 0
		}
		for i := 1; i < len(tokens); i++ {
			if tokens[i].kind == from {
				return 0, i + 1
			}
		}

		return 0, 0
	}
	readContainerAlias := func(tokens []token, i int) (string, int) {
		if tokens[i].kind == ident {
			return tokens[i].value.(string), i + 1
		}
		return "", i
	}
	readWhereClause := func(tokens []token, i int, containerAlias string) ([]Instruction, int) {
		if tokens[i].kind != where {
			return nil, i
		}

		clauses := make([]Instruction, 0, 4)
		levelSep := "/"
		keyBuilder := ""
		op := Eq
		cmd := Push
		for i++; i < len(tokens); i++ {
			if tokens[i] == (token{ident, containerAlias}) {
				i++
			}
			if tokens[i].kind == dot {
				keyBuilder += levelSep
			}
			if tokens[i].kind == ident {
				keyBuilder += tokens[i].value.(string)
			}
			if tokens[i].kind == eq {
				op = Eq
			}
			if tokens[i].kind == gt {
				op = Gt
			}
			if tokens[i].kind == text || tokens[i].kind == float {
				clauses = append(clauses, Instruction{cmd, keyBuilder, op, tokens[i].value})
				keyBuilder = ""
			}
			if tokens[i].kind == and {
				cmd = And
			}
			if tokens[i].kind == or {
				cmd = Or
			}
		}
		return clauses, i
	}

	tokens, err := tokenize(query)
	if err != nil {
		return Program{Instructions: nil}, err
	}
	_, i := readSelection(tokens) // TODO
	containerAlias, i := readContainerAlias(tokens, i)
	instructions, i := readWhereClause(tokens, i, containerAlias)

	return Program{Instructions: instructions}, nil
}
