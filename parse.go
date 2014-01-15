package parser

func ParseString(input string, c Component) (*Result, error) {
	b := NewBuffer(input)
	return c.Parse(b)
}
