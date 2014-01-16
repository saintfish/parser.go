package parser

func ParseString(input string, pos int, c Component) (Value, error) {
	b := newBuffer(input, pos)
	return c.Parse(b)
}
