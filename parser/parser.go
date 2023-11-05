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
