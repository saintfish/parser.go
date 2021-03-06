package parser

import (
	"fmt"
	"github.com/saintfish/trie.go"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Run struct {
	Start, End int
}

type Buffer struct {
	input   string
	pos     int
	backPos []int
}

func newBuffer(input string, pos int) *Buffer {
	if pos > len(input) {
		pos = len(input)
	}
	return &Buffer{
		input:   input,
		pos:     pos,
		backPos: make([]int, 0, 10),
	}
}

func (b *Buffer) Backup() {
	b.backPos = append(b.backPos, b.pos)
}

func (b *Buffer) Restore() error {
	err := fmt.Errorf(
		"parse error at %d, restore to %d",
		b.pos, b.backPos[len(b.backPos)-1])
	b.pos = b.backPos[len(b.backPos)-1]
	b.backPos = b.backPos[:len(b.backPos)-1]
	return err
}

func (b *Buffer) End() bool {
	return b.pos >= len(b.input)
}

func (b *Buffer) Commit() Run {
	run := Run{b.backPos[len(b.backPos)-1], b.pos}
	b.backPos = b.backPos[:len(b.backPos)-1]
	return run
}

func (b *Buffer) Run(r Run) string {
	return b.input[r.Start:r.End]
}

func (b *Buffer) ConsumePrefix(prefix string) bool {
	if strings.HasPrefix(b.input[b.pos:], prefix) {
		b.pos += len(prefix)
		return true
	}
	return false
}

func (b *Buffer) ConsumeRune() (rune, bool) {
	if b.End() {
		return 0, false
	}
	r, size := utf8.DecodeRuneInString(b.input[b.pos:])
	b.pos += size
	return r, true
}

func (b *Buffer) ConsumeTrie(t *trie.Trie) (trie.Value, bool) {
	if m, found := t.MatchLongestPrefixString(b.input[b.pos:]); found {
		b.pos += m.PrefixLength
		return m.Value, true
	}
	return nil, false
}

func (b *Buffer) ConsumeRegexp(r *regexp.Regexp) bool {
	index := r.FindStringIndex(b.input[b.pos:])
	if index == nil {
		return false
	}
	b.pos += (index[1] - index[0])
	return true
}
