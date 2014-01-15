package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	val := Indirect()
	spaces := Regexp("[ ]*")
	number := Regexp("\\d+")
	// <product> := <val> { ("*" | "/") <val> }
	product := Cat(val, Repeat(Cat(spaces, Rune("*/"), spaces, val)))
	// <sum> := <product> { ("+" | "-") <product> }
	sum := Cat(product, Repeat(Cat(spaces, Rune("+-"), spaces, product)))
	// <expr> := <sum>
	expr := sum
	// <val> := <number> | "(" <expr> ")"
	val.Init(Alter(number, Cat(Literal("("), spaces, expr, spaces, Literal(")"))))

	r, err := ParseString("1 + 2 * (3 * 4 + (5)) + 6", expr)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s", r.String())
}
