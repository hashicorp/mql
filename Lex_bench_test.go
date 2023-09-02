package mql

import (
	"math/rand"
	"testing"

	lx "github.com/hashicorp/mql/lexer"
)

var charSet = []rune("01234567890+-*/()")

func randText(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(b)
}

// the current implementation
func BenchmarkLexerStack(b *testing.B) {
	l := newLexer(randText(b.N))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.read()
		l.unread()

		l.read()
		l.read()
		l.read()
		l.unread()

		runesToString(l.current)
		l.current.clear()
	}
}

// the new implementation
func BenchmarkLexerCursor(b *testing.B) {
	l := lx.New(randText(b.N))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Shift()
		l.Backup()

		l.Shift()
		l.Shift()
		l.Shift()
		l.Backup()

		l.Reduce()
	}
}
