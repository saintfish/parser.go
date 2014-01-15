package parser

func ParseString(input string, c Component) (*Result, error) {
	b := NewBuffer([]byte(input))
	return c.Parse(b)
}
