package parser

import (
	"github.com/saintfish/trie.go"
	"regexp"
)

type Value interface{}

type Component interface {
	Parse(buf *Buffer) (Value, error)
}

type SingletonHandler func(buf *Buffer, r Run) Value

func defaultSingletonHandler(buf *Buffer, r Run) Value {
	return r
}

func Literal(prefix string) Component {
	return HandleLiteral(nil, prefix)
}

func HandleLiteral(handler SingletonHandler, prefix string) Component {
	if handler == nil {
		handler = defaultSingletonHandler
	}
	f := func(buf *Buffer) (Value, error) {
		buf.Backup()
		if buf.ConsumePrefix(prefix) {
			return handler(buf, buf.Commit()), nil
		} else {
			return nil, buf.Restore()
		}
	}
	return &generalComponent{f: f}
}

func Rune(chars string) Component {
	return HandleRune(nil, chars)
}

func HandleRune(handler SingletonHandler, chars string) Component {
	if handler == nil {
		handler = defaultSingletonHandler
	}
	f := func(buf *Buffer) (Value, error) {
		buf.Backup()
		if r, ok := buf.ConsumeRune(); ok {
			for _, ch := range chars {
				if r == ch {
					return handler(buf, buf.Commit()), nil
				}
			}
		}
		return nil, buf.Restore()
	}
	return &generalComponent{f: f}
}

func Dict(dict ...string) Component {
	return HandleDict(nil, dict...)
}

func HandleDict(handler SingletonHandler, dict ...string) Component {
	if handler == nil {
		handler = defaultSingletonHandler
	}
	t := trie.NewTrie()
	for _, key := range dict {
		t.Add([]byte(key), nil)
	}
	f := func(buf *Buffer) (Value, error) {
		buf.Backup()
		if _, ok := buf.ConsumeTrie(t); ok {
			return handler(buf, buf.Commit()), nil
		}
		return nil, buf.Restore()
	}
	return &generalComponent{f: f}
}

func HandleTrie(handler func(buf *Buffer, r Run, v trie.Value) Value, t *trie.Trie) Component {
	if handler == nil {
		handler = func(buf *Buffer, r Run, v trie.Value) Value {
			return v
		}
	}
	f := func(buf *Buffer) (Value, error) {
		buf.Backup()
		if vv, ok := buf.ConsumeTrie(t); ok {
			return handler(buf, buf.Commit(), vv), nil
		}
		return nil, buf.Restore()
	}
	return &generalComponent{f: f}
}

func Regexp(re string) Component {
	return HandleRegexp(nil, re)
}

func HandleRegexp(handler SingletonHandler, re string) Component {
	if handler == nil {
		handler = defaultSingletonHandler
	}
	if re[0] != '^' {
		re = "^" + re
	}
	reg := regexp.MustCompile(re)
	f := func(buf *Buffer) (Value, error) {
		buf.Backup()
		if buf.ConsumeRegexp(reg) {
			return handler(buf, buf.Commit()), nil
		}
		return nil, buf.Restore()
	}
	return &generalComponent{f: f}
}

type CompositeHandler func(b *Buffer, r Run, values []Value) Value

func defaultCompositeHandler(b *Buffer, r Run, values []Value) Value {
	return values
}

func Repeat(c Component) Component {
	return HandleRepeat(nil, c)
}

func HandleRepeat(handler CompositeHandler, c Component) Component {
	if handler == nil {
		handler = defaultCompositeHandler
	}
	f := func(buf *Buffer) (Value, error) {
		buf.Backup()
		values := []Value{}
		for {
			buf.Backup()
			value, err := c.Parse(buf)
			if err == nil {
				buf.Commit()
				values = append(values, value)
			} else {
				buf.Restore()
				break
			}
		}
		return handler(buf, buf.Commit(), values), nil
	}
	return &generalComponent{f: f}
}

func Option(c Component) Component {
	return HandleOption(nil, c)
}

func HandleOption(handler func(*Buffer, Run, Value) Value, c Component) Component {
	if handler == nil {
		handler = func(buf *Buffer, r Run, v Value) Value {
			return v
		}
	}
	f := func(buf *Buffer) (Value, error) {
		buf.Backup()
		value, err := c.Parse(buf)
		if err != nil {
			buf.Restore()
			return nil, nil
		}
		return handler(buf, buf.Commit(), value), nil
	}
	return &generalComponent{f: f}
}

func Cat(components ...Component) Component {
	return HandleCat(nil, components...)
}

func HandleCat(handler CompositeHandler, components ...Component) Component {
	if handler == nil {
		handler = defaultCompositeHandler
	}
	f := func(buf *Buffer) (Value, error) {
		buf.Backup()
		values := []Value{}
		for _, c := range components {
			value, err := c.Parse(buf)
			if err != nil {
				return nil, buf.Restore()
			}
			values = append(values, value)
		}
		return handler(buf, buf.Commit(), values), nil
	}
	return &generalComponent{f: f}
}

func Alter(alts ...Component) Component {
	return HandleAlter(nil, alts...)
}

func HandleAlter(handler func(b *Buffer, r Run, v Value, index int) Value, alts ...Component) Component {
	if handler == nil {
		handler = func(b *Buffer, r Run, v Value, index int) Value {
			return v
		}
	}
	f := func(buf *Buffer) (Value, error) {
		var err error
		var value Value
		for i, c := range alts {
			buf.Backup()
			value, err = c.Parse(buf)
			if err == nil {
				return handler(buf, buf.Commit(), value, i), nil
			}
			err = buf.Restore()
		}
		return nil, err
	}
	return &generalComponent{f: f}
}

type IndirectComponent struct {
	c Component
}

func Indirect() *IndirectComponent {
	return &IndirectComponent{}
}

func (c *IndirectComponent) Init(cc Component) {
	c.c = cc
}

func (c *IndirectComponent) Parse(buf *Buffer) (Value, error) {
	if c.c == nil {
		panic("indirect component was used before initialized")
	}
	return c.c.Parse(buf)
}

type generalComponent struct {
	f func(buf *Buffer) (Value, error)
}

func (c *generalComponent) Parse(buf *Buffer) (Value, error) {
	return c.f(buf)
}
