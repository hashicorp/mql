package lexer

import (
	"errors"
	"unicode/utf8"
)

func New(text string) *Lexer {
	return &Lexer{buf: text}
}

type Lexer struct {
	buf      string
	off      int
	lastRead readOp
	pos      int
	eof      bool
}

// used to track the size of the last read rune, since ut8 chars can be bigger
// than one byte
type readOp int8

const (
	opRead readOp = iota - 1
	opInvalid
	opReadRune1
	opReadRune2
	opReadRune3
	opReadRune4
)

const (
	RuneErr rune = -1
	RuneEOF rune = 0
)

// true, if the offset is at the end of the input buffer
func (l *Lexer) empty() bool { return len(l.buf) <= l.off }

// length of unread portion of the input buffer
func (l *Lexer) Len() int { return len(l.buf) - l.off }

// offset from the start of the input buffer
func (l *Lexer) Off() int { return l.off }

func (l *Lexer) Diff(v string) int { return l.off - len(v) }

// return the next rune.
// if the input is empty, a synthetic
// EOF rune, with value 0, is emitted and
// the offset is not modified
func (l *Lexer) Shift() rune {
	if l.empty() {
		// emit artificial eof rune
		l.eof = true
		l.lastRead = opReadRune1
		return RuneEOF
	}
	c := l.buf[l.off]
	if c < utf8.RuneSelf {
		l.off++
		l.lastRead = opReadRune1
		return rune(c)
	}
	r, n := utf8.DecodeRuneInString(l.buf[l.off:])
	l.off += n
	l.lastRead = readOp(n)
	return r
}

// move the offset back by the size of the last read rune.
// only 1 backup is possible.
// if the last rune was EOF, the offset is not actually modified
func (l *Lexer) Backup() error {
	if l.lastRead <= opInvalid {
		return errors.New("UnreadRune: previous operation was not a successful ReadRune")
	}
	// account for artificial eof rune
	if l.eof {
		l.eof = false
		return nil
	}
	if l.off >= int(l.lastRead) {
		l.off -= int(l.lastRead)
	}
	// this prevents to backup twice, because the size of
	// the previous rune, is not known
	l.lastRead = opInvalid
	return nil
}

// get the runes from the last call of Reduce
// to the offset.
func (l *Lexer) Reduce() string {
	v := l.buf[l.pos:l.off]
	l.pos = l.off
	return string(v)
}

// get the next rune, without mutating the offset
func (l *Lexer) Peek() rune {
	r := l.Shift()
	l.Backup()
	return r
}

// advance, if the next rune passes the check
func (l *Lexer) Expect(valid CheckFn) bool {
	if !valid(l.Shift()) {
		l.Backup()
		return false
	}
	return true
}

// advance, for as long as subsequent runes
// pass the check. Returns false if not at least 1
// rune has been consumed
func (l *Lexer) Some(valid CheckFn) bool {
	if !valid(l.Shift()) {
		l.Backup()
		return false
	}
	for valid(l.Shift()) {
	}
	l.Backup()
	return true
}
