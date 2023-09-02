package lexer

import (
	"strings"
	"unicode"
)

// check if a given rune matches a given criteria
type CheckFn func(rune) bool

var (
	IsEOF                = Eq(RuneEOF)
	IsSpace              = unicode.IsSpace
	IsNumber             = unicode.IsDigit
	IsLetter             = unicode.IsLetter
	IsDoubleQuote        = Eq('"')
	IsArithmeticOperator = In("+-*/")
	IsLogicSymbol        = In("&|")
	IsParenthesisLeft    = Eq('(')
	IsParenthesisRight   = Eq(')')
	IsEQ                 = Eq('=')
)

func Eq(valid rune) CheckFn {
	return func(r rune) bool { return r == valid }
}

func In(valid string) CheckFn {
	return func(r rune) bool { return strings.ContainsRune(valid, r) }
}

func Not(valid CheckFn) CheckFn {
	return func(r rune) bool { return !valid(r) }
}

func Or(checks ...CheckFn) CheckFn {
	return func(r rune) bool {
		for _, valid := range checks {
			if valid(r) {
				return true
			}
		}
		return false
	}
}

func And(checks ...CheckFn) CheckFn {
	return func(r rune) bool {
		for _, valid := range checks {
			if !valid(r) {
				return false
			}
		}
		return true
	}
}
