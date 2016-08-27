package pre

import (
	"bytes"
	"io/ioutil"
	"log"
	"unicode"
)

type readBuffer struct {
	buffer   []byte
	runes    []rune
	position int
}

func (b *readBuffer) readFile(name string) {
	buffer, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	b.buffer = buffer
	b.runes = bytes.Runes(b.buffer)
}

func (b *readBuffer) readText(text string) {
	b.buffer = []byte(text)
	b.runes = bytes.Runes(b.buffer)
}

func (b *readBuffer) current() rune {
	if b.isEnd() {
		return unicode.MaxRune
	}
	return b.runes[b.position]
}

func (b *readBuffer) next() rune {
	if b.isEnd() {
		return unicode.MaxRune
	}
	b.position++
	return b.runes[b.position-1]
}

func (b *readBuffer) push() {
	b.position--
}

func (b *readBuffer) peek() (rune, bool) {
	end := b.isEnd()
	if end {
		return unicode.MaxRune, true
	}
	return b.runes[b.position+1], false
}

func (b *readBuffer) isEnd() bool {
	return len(b.runes) <= b.position
}
