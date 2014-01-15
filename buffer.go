package parser

import (
	"bytes"
	"fmt"
	"github.com/saintfish/trie.go"
	"regexp"
	"unicode/utf8"
)

type Run struct {
	Start, End int
}

type Buffer struct {
	input   []byte
	pos     int
	backPos []int
}

func NewBuffer(input []byte) *Buffer {
	return &Buffer{
		input:   input,
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

func (b *Buffer) ConsumePrefix(prefix []byte) bool {
	if b.End() {
		return false
	}
	if bytes.HasPrefix(b.input[b.pos:], prefix) {
		b.pos += len(prefix)
		return true
	}
	return false
}

func (b *Buffer) ConsumeRune() (rune, bool) {
	if b.End() {
		return 0, false
	}
	r, size := utf8.DecodeRune(b.input[b.pos:])
	b.pos += size
	return r, true
}

func (b *Buffer) ConsumeTrie(t *trie.Trie) bool {
	if b.End() {
		return false
	}
	if m, found := t.MatchLongestPrefix(b.input[b.pos:]); found {
		b.pos += len(m.Prefix)
		return true
	}
	return false
}

func (b *Buffer) ConsumeRegexp(r *regexp.Regexp) bool {
	if b.End() {
		return false
	}
	match := r.Find(b.input[b.pos:])
	if match == nil {
		return false
	}
	b.pos += len(match)
	return true
}
