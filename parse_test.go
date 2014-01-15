package parser

import (
	"fmt"
	"strconv"
	"testing"
)

func TestParse(t *testing.T) {
	val := Indirect()
	spaces := Regexp("[ ]*")
	number := HandleRegexp(
		func(b *Buffer, run Run) Value {
			i, _ := strconv.Atoi(b.Run(run))
			return float32(i)
		},
		"\\d+")
	// <product> := <val> { ("*" | "/") <val> }
	product := HandleCat(
		func(b *Buffer, run Run, values []Value) Value {
			n := values[0].(float32)
			for _, opValueWrapper := range values[1].([]Value) {
				opValue := opValueWrapper.([]Value)
				op := b.Run(opValue[1].(Run))
				val := opValue[3].(float32)
				switch op {
				case "*":
					n *= val
				case "/":
					n /= val
				}
			}
			return n
		},
		val, Repeat(Cat(spaces, Rune("*/"), spaces, val)))
	// <sum> := <product> { ("+" | "-") <product> }
	sum := HandleCat(
		func(b *Buffer, run Run, values []Value) Value {
			n := values[0].(float32)
			for _, opValueWrapper := range values[1].([]Value) {
				opValue := opValueWrapper.([]Value)
				op := b.Run(opValue[1].(Run))
				val := opValue[3].(float32)
				switch op {
				case "+":
					n += val
				case "-":
					n -= val
				}
			}
			return n
		},
		product, Repeat(Cat(spaces, Rune("+-"), spaces, product)))
	// <expr> := <sum>
	expr := sum
	// <val> := <number> | "(" <expr> ")"
	val.Init(HandleAlter(
		func(b *Buffer, r Run, v Value, index int) Value {
			if index == 0 {
				return v
			}
			return v.([]Value)[2]
		},
		number, Cat(Literal("("), spaces, expr, spaces, Literal(")"))))

	input := "1 + 2 * (3 * 4 + (5)) + 6"
	v, err := ParseString(input, expr)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s = %v\n", input, v)
	// 1 + 2 * (3 * 4 + (5)) + 6 = 41
}
