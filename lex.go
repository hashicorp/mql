// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

type lexStateFunc func(*lexer) (lexStateFunc, error)

type lexer struct {
	source  *bufio.Reader
	current stack[rune]
	tokens  chan token
	state   lexStateFunc
}

func newLexer(s string) *lexer {
	l := &lexer{
		source: bufio.NewReader(strings.NewReader(s)),
		state:  lexStartState,
		tokens: make(chan token, 1), // define a ring buffer for emitted tokens
	}
	return l
}

// nextToken is the external api for the lexer and it simply returns the next
// token or an error. If EOF is encountered while scanning, nextToken will keep
// returning an eofToken no matter how many times you call nextToken.
func (l *lexer) nextToken() (token, error) {
	for {
		select {
		case tk := <-l.tokens: // return a token if one has been emitted
			return tk, nil
		default: // otherwise, keep scanning via the next state
			var err error
			if l.state, err = l.state(l); err != nil {
				return token{}, err
			}

		}
	}
}

// lexStartState  is the start state.  It doesn't emit tokens, but rather
// transitions to other states.  Other states typically transition back to
// lexStartState after they emit a token.
func lexStartState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexStartState", "lexer")
	r := l.read()
	switch {
	// wait, if it's eof we're done
	case r == eof:
		l.emit(eofToken, "")
		return lexEofState, nil

	// start with finding all tokens that can have a trailing "="
	case r == '>':
		return lexGreaterState, nil
	case r == '<':
		return lexLesserState, nil

		// now, we can just look at the next rune...
	case r == '%':
		return lexContainsState, nil
	case r == '=':
		return lexEqualState, nil
	case r == '!':
		return lexNotEqualState, nil
	case r == ')':
		return lexRightParenState, nil
	case r == '(':
		return lexLeftParenState, nil
	case isSpace(r):
		return lexWhitespaceState, nil
	default:
		l.unread()
		return lexStringState, nil
	}
}

// lexStringState scans for strings and can emit the following tokens:
// orToken, andToken, containsToken, stringToken
func lexStringState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexStringState", "lexer")
	defer l.current.clear()

	// we'll push the runes we read into this buffer and when appropriate will
	// emit tokens using the buffer's data.
	var buf bytes.Buffer

	// before we start looping, let's found out if we're scanning a quoted string
	r := l.read()
	var quotedString bool
	switch r {
	case '"':
		quotedString = true
	default:
		l.unread()
	}

WriteToBuf:
	// keep reading runes into the buffer until we encounter eof of non-text runes.
	for {
		r = l.read()
		switch {
		case r == eof:
			break WriteToBuf
		case r == '"' && quotedString: // end of the quoted string we're scanning
			break WriteToBuf
		case (isSpace(r) || isSpecial(r)) && !quotedString: // whitespace or a special char, and we're not scanning a quoted string
			l.unread()
			break WriteToBuf
		default: // otherwise, write the rune into the keyword buffer
			buf.WriteRune(r)
		}
	}

	// before emitting a token, do we have a special string?  But, first let's
	// check if we're dealing with a quoted string, since we want to support
	// emitting string tokens for "and", "or" so those tokens can be used in
	// comparison expr. Example: name % "Johnson and"
	if !quotedString {
		switch strings.ToLower(buf.String()) {
		case "and":
			l.emit(andToken, "and")
			return lexStartState, nil
		case "or":
			l.emit(orToken, "or")
			return lexStartState, nil
		}
	}

	l.emit(stringToken, buf.String())
	return lexStartState, nil
}

// lexContainsState emits an containsToken and returns to the lexStartState
func lexContainsState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexContainsState", "lexer")
	defer l.current.clear()
	l.emit(containsToken, "%")
	return lexStartState, nil
}

// lexEqualState emits an equalToken and returns to the lexStartState
func lexEqualState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexEqualState", "lexer")
	defer l.current.clear()
	l.emit(equalToken, "=")
	return lexStartState, nil
}

// lexNotEqualState scans for a notEqualToken and return either to the lexStartState or
// lexErrorState
func lexNotEqualState(l *lexer) (lexStateFunc, error) {
	const op = "mql.lexNotEqualState"
	panicIfNil(l, "lexNotEqualState", "lexer")
	defer l.current.clear()
	nextRune := l.read()
	switch nextRune {
	case '=':
		l.emit(notEqualToken, "!=")
		return lexStartState, nil
	default:
		return nil, fmt.Errorf("%s: %w, got %q", op, ErrInvalidNotEqual, fmt.Sprintf("%s%s", "!", string(nextRune)))
	}
}

// lexLeftParenState emits a startLogicalExprToken and returns to the
// lexStartState
func lexLeftParenState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexLeftParenState", "lexer")
	defer l.current.clear()
	l.emit(startLogicalExprToken, runesToString(l.current))
	return lexStartState, nil
}

// lexRightParenState emits an endLogicalExprToken and returns to the
// lexStartState
func lexRightParenState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexRightParenState", "lexer")
	defer l.current.clear()
	l.emit(endLogicalExprToken, runesToString(l.current))
	return lexStartState, nil
}

// lexWhitespaceState emits a whitespaceToken and returns to the lexStartState
func lexWhitespaceState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexWhitespaceState", "lexer")
	defer l.current.clear()
ReadWhitespace:
	for {
		ch := l.read()
		switch {
		case ch == eof:
			break ReadWhitespace
		case !isSpace(ch):
			l.unread()
			break ReadWhitespace
		}
	}
	l.emit(whitespaceToken, "")
	return lexStartState, nil
}

// lexGreaterState will emit either a greaterThanToken or a
// greaterThanOrEqualToken and return to the lexStartState
func lexGreaterState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexGreaterState", "lexer")
	defer l.current.clear()
	next := l.read()
	switch next {
	case '=':
		l.emit(greaterThanOrEqualToken, ">=")
		return lexStartState, nil
	default:
		l.unread()
		l.emit(greaterThanToken, ">")
		return lexStartState, nil
	}
}

// lexLesserState will emit either a lessThanToken or a lessThanOrEqualToken and
// return to the lexStartState
func lexLesserState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexLesserState", "lexer")
	defer l.current.clear()
	next := l.read()
	switch next {
	case '=':
		l.emit(lessThanOrEqualToken, "<=")
		return lexStartState, nil
	default:
		l.unread()
		l.emit(lessThanToken, "<")
		return lexStartState, nil
	}
}

// lexEofState will emit an eofToken and returns right back to the lexEofState
func lexEofState(l *lexer) (lexStateFunc, error) {
	panicIfNil(l, "lexEofState", "lexer")
	l.emit(eofToken, "")
	return lexEofState, nil
}

// emit send a token to the lexer's token channel
func (l *lexer) emit(t tokenType, v string) {
	l.tokens <- token{
		Type:  t,
		Value: v,
	}
}

// isSpace reports if r is a space
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

// isSpecial reports r is special rune
func isSpecial(r rune) bool {
	return r == '=' || r == '>' || r == '!' || r == '<' || r == '(' || r == ')' || r == '%'
}

// read the next rune
func (l *lexer) read() rune {
	ch, _, err := l.source.ReadRune()
	if err != nil {
		return eof
	}
	l.current.push(ch)
	return ch
}

// unread the last rune read which means that rune will be returned the next
// time lexer.read() is called.  unread also removes the last rune from the
// lexer's stack of current runes
func (l *lexer) unread() {
	_ = l.source.UnreadRune() // error ignore which only occurs when nothing has been previously read
	_, _ = l.current.pop()
}
