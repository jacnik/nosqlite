package parser

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
	and
)

type token struct {
	tokenType tokenType
	value     any
}

func tokenize(query string) []token {
	return []token{
		{sel, nil},
		{star, nil},
		token{from, nil},
		token{ident, "c"},
		token{where, nil},
		token{ident, "c"},
		token{dot, nil},
		token{ident, "social"},
		token{dot, nil},
		token{ident, "twitter"},
		token{eq, nil},
		token{eq, nil},
		token{text, "https://twitter.com"},
		token{and, nil},
		token{ident, "c"},
		token{dot, nil},
		token{ident, "age"},
		token{gt, nil},
		token{float, 17.0},
	}

}

// func tokenize(query string) []token {
// 	min := func(x, y int) int {
// 		if y < x {
// 			return y
// 		}
// 		return x
// 	}

// 	isWhitespace := func(query string, i int) bool {
// 		spaces := []byte{' ', '\t', '\n'}
// 		for _, s := range spaces {
// 			if query[i] == s {
// 				return true
// 			}
// 		}
// 		return false
// 	}

// 	isSpecialChar := func(query string, i int) (tokenType, bool) {
// 		if i >= len(query) {
// 			return eof, true
// 		}

// 		switch query[i] {
// 		case '(':
// 			return lparem, true
// 		case ')':
// 			return rparem, true
// 		case '[':
// 			return lsqbrack, true
// 		case ']':
// 			return rsqbrack, true
// 		case '.':
// 			return dot, true
// 		case ',':
// 			return comma, true
// 		case '*':
// 			return star, true
// 		}

// 		return eof, false
// 	}

// 	isTextChar := func(query string, i int) (newPos int, ok bool) {
// 		c := query[i]
// 		if c >= 'a' && c <= 'z' {
// 			return i + 1, true
// 		}
// 		if c >= 'A' && c <= 'Z' {
// 			return i + 1, true
// 		}
// 		if c == '_' || c == '-' {
// 			return i + 1, true
// 		}
// 		return i, false
// 	}

// 	isDigitChar := func(query string, i int) (newPos int, ok bool) {
// 		c := query[i]
// 		if c >= '0' && c <= '9' {
// 			return i + 1, true
// 		}
// 		if c == '.' || c == '-' {
// 			return i + 1, true
// 		}
// 		return i, false
// 	}

// 	passWhitespace := func(query string, i int) int {
// 		j := i
// 		for ; j < len(query); j++ {
// 			if !isWhitespace(query, j) {
// 				return j
// 			}
// 		}
// 		return j
// 	}
// 	readSelect := func(query string, i int) (int, *token) {
// 		s := "SELECT"
// 		l := len(s)

// 		if query[i:min(len(query)-i, l)] == s {
// 			return passWhitespace(query, i+l), &token{sel, s}
// 		}
// 		return i, nil
// 	}

// 	readSelections := func(query string, i int) (int, []token) {
// 		res := make([]token, 0, 1)
// 		i = passWhitespace(query, i)
// 		if query[i] == '*' {
// 			res = append(res, token{tokenType: star})
// 			return passWhitespace(query, i+1), res
// 		}
// 		panic("selectors other than '*' are not supported yet.") // TODO
// 	}

// 	readFrom := func(query string, i int) (int, *token) {
// 		s := "FROM"
// 		l := len(s)

// 		if query[i:min(len(query)-i, l)] == s {
// 			return passWhitespace(query, i+l), &token{sel, s}
// 		}
// 		return i, nil
// 	}

// 	readText := func(query string, i int) (int, *token) {
// 		i = passWhitespace(query, i)

// 		for j := i; j < len(query); j++ {
// 			if isWhitespace(query, j) {
// 				return passWhitespace(query, j+1), &token{text, query[i:j]}
// 			}
// 		}

// 		return len(query), &token{eof, nil}
// 	}

// 	/* SELECT * FROM c WHERE c.social.twitter = 'https://twitter.com' AND c.social.facebook = 'https://facebook.com' */

// 	tokens := make([]token, 0, 8)

// 	for s := 0; s < len(query); {
// 		i, t := readSelect(query, s) // SELECT
// 		if t != nil {
// 			tokens = append(tokens, *t)
// 		}
// 		i, ts := readSelections(query, i) // *
// 		tokens = append(tokens, ts...)
// 		i, t = readFrom(query, s) // FROM
// 		tokens = append(tokens, *t)
// 		i, t = readText(query, s) // c
// 		tokens = append(tokens, *t)

// 		// WHERE
// 		// c.social.twitter = 'https://twitter.com' AND c.social.facebook = 'https://facebook.com'

// 	}

// 	return tokens
// }
