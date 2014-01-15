package parser

import (
	"bytes"
	"fmt"
	"github.com/saintfish/trie.go"
	"regexp"
	"strings"
)

type Result struct {
	Run      Run
	Children []*Result
}

func (r *Result) String() string {
	b := &bytes.Buffer{}
	r.stringIndent(0, b)
	return b.String()
}

func (r *Result) stringIndent(indent int, b *bytes.Buffer) {
	if r == nil {
		b.WriteString(fmt.Sprintf("%snil\n", strings.Repeat("\t", indent)))
		return
	}
	b.WriteString(fmt.Sprintf("%s%d %d\n", strings.Repeat("\t", indent), r.Run.Start, r.Run.End))
	for _, c := range r.Children {
		c.stringIndent(indent+1, b)
	}
}

type Component interface {
	Parse(buf *Buffer) (*Result, error)
}

func Literal(s string) Component {
	prefix := []byte(s)
	f := func(buf *Buffer) (*Result, error) {
		buf.Backup()
		if buf.ConsumePrefix(prefix) {
			return &Result{Run: buf.Commit()}, nil
		} else {
			return nil, buf.Restore()
		}
	}
	return &generalComponent{f: f}
}

func Rune(chars string) Component {
	f := func(buf *Buffer) (*Result, error) {
		buf.Backup()
		if r, ok := buf.ConsumeRune(); ok {
			for _, ch := range chars {
				if r == ch {
					return &Result{Run: buf.Commit()}, nil
				}
			}
		}
		return nil, buf.Restore()
	}
	return &generalComponent{f: f}
}

func Dict(dict []string) Component {
	t := trie.NewTrie()
	for _, key := range dict {
		t.Add([]byte(key), nil)
	}
	f := func(buf *Buffer) (*Result, error) {
		buf.Backup()
		if buf.ConsumeTrie(t) {
			return &Result{Run: buf.Commit()}, nil
		}
		return nil, buf.Restore()
	}
	return &generalComponent{f: f}
}

func Regexp(re string) Component {
	if re[0] != '^' {
		re = "^" + re
	}
	reg := regexp.MustCompile(re)
	f := func(buf *Buffer) (*Result, error) {
		buf.Backup()
		if buf.ConsumeRegexp(reg) {
			return &Result{Run: buf.Commit()}, nil
		}
		return nil, buf.Restore()
	}
	return &generalComponent{f: f}
}

func Repeat(c Component) Component {
	f := func(buf *Buffer) (*Result, error) {
		buf.Backup()
		allResult := &Result{}
		for {
			buf.Backup()
			result, err := c.Parse(buf)
			if err == nil {
				buf.Commit()
				allResult.Children = append(allResult.Children, result)
			} else {
				buf.Restore()
				break
			}
		}
		allResult.Run = buf.Commit()
		return allResult, nil
	}
	return &generalComponent{f: f}
}

func Option(c Component) Component {
	f := func(buf *Buffer) (*Result, error) {
		buf.Backup()
		result, err := c.Parse(buf)
		if err != nil {
			buf.Restore()
			return nil, nil
		}
		buf.Commit()
		return result, nil
	}
	return &generalComponent{f: f}
}

func Cat(components ...Component) Component {
	f := func(buf *Buffer) (*Result, error) {
		buf.Backup()
		allResult := &Result{}
		for _, c := range components {
			result, err := c.Parse(buf)
			if err != nil {
				return nil, buf.Restore()
			}
			allResult.Children = append(allResult.Children, result)
		}
		allResult.Run = buf.Commit()
		return allResult, nil
	}
	return &generalComponent{f: f}
}

func Alter(alts ...Component) Component {
	f := func(buf *Buffer) (*Result, error) {
		var err error
		var result *Result
		for _, c := range alts {
			buf.Backup()
			result, err = c.Parse(buf)
			if err == nil {
				buf.Commit()
				return result, nil
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

func (c *IndirectComponent) Parse(buf *Buffer) (*Result, error) {
	if c.c == nil {
		panic("indirect component was used before initialized")
	}
	return c.c.Parse(buf)
}

type generalComponent struct {
	f func(buf *Buffer) (*Result, error)
}

func (c *generalComponent) Parse(buf *Buffer) (*Result, error) {
	return c.f(buf)
}
