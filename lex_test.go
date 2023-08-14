// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_lexKeywordState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		raw             string
		want            []token
		wantErrIs       error
		wantErrContains string
	}{
		{
			name: "just-eof",
			raw:  ``,
			want: []token{
				{Type: eofToken, Value: ""},
				{Type: eofToken, Value: ""}, // will continue until you stop calling lexer.NextToken
			},
		},
		{
			name: "empty-quotes",
			raw:  `name=""`,
			want: []token{
				{Type: stringToken, Value: "name"},
				{Type: equalToken, Value: "="},
				{Type: stringToken, Value: ""},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "quoted-value",
			raw:  `"value"`,
			want: []token{
				{Type: stringToken, Value: `value`},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "non-quoted-value",
			raw:  "non-quoted-value",
			want: []token{
				{Type: stringToken, Value: "non-quoted-value"},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "comparison-op-in-keyword",
			raw:  "greater>than",
			want: []token{
				{Type: stringToken, Value: "greater"},
				{Type: greaterThanToken, Value: ">"},
				{Type: stringToken, Value: "than"},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "%",
			raw:  "%",
			want: []token{
				{Type: containsToken, Value: "%"},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "and",
			raw:  "and ",
			want: []token{
				{Type: andToken, Value: "and"},
				{Type: whitespaceToken, Value: ""},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "or",
			raw:  "or ",
			want: []token{
				{Type: orToken, Value: "or"},
				{Type: whitespaceToken, Value: ""},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "greaterThan",
			raw:  ">",
			want: []token{
				{Type: greaterThanToken, Value: ">"},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "greaterThanOrEqual",
			raw:  ">=",
			want: []token{
				{Type: greaterThanOrEqualToken, Value: ">="},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "lessThan",
			raw:  "<",
			want: []token{
				{Type: lessThanToken, Value: "<"},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "lessThanOrEqual",
			raw:  "<=",
			want: []token{
				{Type: lessThanOrEqualToken, Value: "<="},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "equal",
			raw:  "=",
			want: []token{
				{Type: equalToken, Value: "="},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "notEqual",
			raw:  "!=",
			want: []token{
				{Type: notEqualToken, Value: "!="},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "notEqualError",
			raw:  "!not",
			want: []token{
				{Type: errToken, Value: `mql.lexNotEqualState: unexpected "=" after "!"`},
				{Type: errToken, Value: `mql.lexNotEqualState: unexpected "=" after "!"`}, // will continue until you stop calling lexer.NextToken
			},
			wantErrIs:       ErrInvalidNotEqual,
			wantErrContains: `mql.lexNotEqualState: invalid "!=" token, got "!n"`,
		},
		{
			name: "startLogicalExpr",
			raw:  "(",
			want: []token{
				{Type: startLogicalExprToken, Value: "("},
				{Type: eofToken, Value: ""},
			},
		},
		{
			name: "endLogicalExpr",
			raw:  ")",
			want: []token{
				{Type: endLogicalExprToken, Value: ")"},
				{Type: eofToken, Value: ""},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)

			lex := newLexer(tc.raw)
			for _, want := range tc.want {
				tk, err := lex.nextToken()
				if tc.wantErrContains != "" {
					require.Error(err)
					if tc.wantErrIs != nil {
						assert.ErrorIs(err, tc.wantErrIs)
					}
					assert.ErrorContains(err, tc.wantErrContains)
					return
				}
				require.NoError(err)
				require.NotEqualValues(tk,
					whitespaceToken,
					startLogicalExprToken,
					endLogicalExprToken,
					greaterThanOp,
					greaterThanOrEqualOp,
					lessThanOp,
					lessThanOrEqualOp,
					equalOp,
					notEqualOp,
					containsOp,
				)
				assert.Equal(want, tk)
			}
		})
	}

}

func Test_lexWhitespaceState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
		want token
	}{
		{
			name: "leading-whitespace",
			raw:  "      leading",
			want: token{
				Type:  whitespaceToken,
				Value: "",
			},
		},
		{
			name: "trailing-whitespace",
			raw:  "  ",
			want: token{
				Type:  whitespaceToken,
				Value: "",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)

			lex := newLexer(tc.raw)
			tk, err := lex.nextToken()
			require.NoError(err)
			require.NotEqualValues(tk.Type,
				errToken,
				eofToken,
				stringToken,
				startLogicalExprToken,
				endLogicalExprToken,
				greaterThanOp,
				greaterThanOrEqualOp,
				lessThanOp,
				lessThanOrEqualOp,
				equalOp,
				notEqualOp,
				containsOp,
			)
			assert.Equal(tc.want, tk)
		})
	}

}

// Fuzz_lexerNextToken is only focused on finding panics
func Fuzz_lexerNextToken(f *testing.F) {
	tc := []string{">=!=", "string ( ) > >=", "< <= = != AND OR and or", "1  !=   \"2\""}
	for _, tc := range tc {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, s string) {
		helperFn := func(lex *lexer) []token {
			var tokens []token
			for {
				tok, err := lex.nextToken()
				if err != nil {
					return tokens
				}
				tokens = append(tokens, tok)
				if tok.Type == eofToken {
					return tokens
				}
			}
		}
		lex := newLexer(s)
		tokens := helperFn(lex)
		for _, token := range tokens {
			if token.Type.String() == "Unknown" {
				t.Errorf("unexpected token %v", token)
			}
		}
	})
}
