package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	spaces := Regexp("[ ]*")
	number := Regexp("\\d+")
	val := Indirect()
	product := Cat(val, Repeat(Cat(spaces, Rune("*/"), spaces, val)))
	sum := Cat(product, Repeat(Cat(spaces, Rune("+-"), spaces, product)))
	expr := sum
	val.Init(Alter(number, Cat(Literal("("), spaces, expr, spaces, Literal(")"))))

	r, err := ParseString("1 + 2 * (3 + 4) + 5", expr)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s", r.String())
}
