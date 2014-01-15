package parser

func ParseString(input string, c Component) (Value, error) {
	b := NewBuffer(input)
	return c.Parse(b)
}
