// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		raw             string
		want            expr
		wantErrIs       error
		wantErrContains string
	}{
		{
			name: "success-comparisonExpr",
			raw:  "name=alice",
			want: &comparisonExpr{
				column:       "name",
				comparisonOp: "=",
				value:        pointer("alice"),
			},
		},
		{
			name: "success-comparisonExpr-with-whitespace",
			raw:  "name= 	alice",
			want: &comparisonExpr{
				column:       "name",
				comparisonOp: "=",
				value:        pointer("alice"),
			},
		},
		{
			name: "success-comparisonExpr-with-parens",
			raw:  "(name=alice)",
			want: &comparisonExpr{
				column:       "name",
				comparisonOp: "=",
				value:        pointer("alice"),
			},
		},
		{
			name: "success-case-sensitive",
			raw:  "FirstName=alice",
			want: &comparisonExpr{
				column:       "FirstName",
				comparisonOp: "=",
				value:        pointer("alice"),
			},
		},
		{
			name: "success-quoted-value",
			raw:  "name!=\"alice eve\"",
			want: &comparisonExpr{
				column:       "name",
				comparisonOp: "!=",
				value:        pointer("alice eve"),
			},
		},
		{
			name: "success-quoted-empty-value",
			raw:  "(name!=\"\" and description=eve) or (name=alice)",
			want: &logicalExpr{
				leftExpr: &logicalExpr{
					leftExpr: &comparisonExpr{
						column:       "name",
						comparisonOp: "!=",
						value:        pointer(""),
					},
					logicalOp: "and",
					rightExpr: &comparisonExpr{
						column:       "description",
						comparisonOp: "=",
						value:        pointer("eve"),
					},
				},
				logicalOp: "or",
				rightExpr: &comparisonExpr{
					column:       "name",
					comparisonOp: "=",
					value:        pointer("alice"),
				},
			},
		},
		{
			name: "success-or-comparison",
			raw:  "name=alice or version >= 110",
			want: &logicalExpr{
				leftExpr: &comparisonExpr{
					column:       "name",
					comparisonOp: "=",
					value:        pointer("alice"),
				},
				logicalOp: "or",
				rightExpr: &comparisonExpr{
					column:       "version",
					comparisonOp: ">=",
					value:        pointer("110"),
				},
			},
		},
		{
			name: "success-quoted-and-emits-string",
			raw:  `name%"and"`,
			want: &comparisonExpr{
				column:       "name",
				comparisonOp: "%",
				value:        pointer("and"),
			},
		},
		{
			name: "success-quoted-or-emits-string",
			raw:  `name="or"`,
			want: &comparisonExpr{
				column:       "name",
				comparisonOp: "=",
				value:        pointer("or"),
			},
		},
		{
			name:            "err-missing-logicalOp",
			raw:             "name=alice (name=eve)",
			wantErrIs:       ErrMissingLogicalOp,
			wantErrContains: "missing logical operator before right side",
		},
		{
			name:            "err-too-many-rightExprs",
			raw:             "name=alice=bob",
			wantErrIs:       ErrUnexpectedToken,
			wantErrContains: `unexpected token eq:"=" in: name=alice=bob`,
		},
		{
			name:            "err-invalid-comparison-expr",
			raw:             "name=alice(invalid=comparison)",
			wantErrIs:       ErrUnexpectedOpeningParen,
			wantErrContains: `unexpected opening paren after (comparisonExpr: name = alice) in: "name=alice(invalid=comparison)`,
		},
		{
			name:            "err-missing-comparison-op",
			raw:             "name",
			wantErrIs:       ErrMissingComparisonOp,
			wantErrContains: "missing comparison operator in: \"name\"",
		},
		{
			name:            "err-missing-comparison-op-in-logical-expr",
			raw:             "name=alice or age",
			wantErrIs:       ErrMissingComparisonOp,
			wantErrContains: "missing comparison operator in: \"name=alice or age\"",
		},
		{
			name:            "err-trailing-logical-op",
			raw:             "name=alice or",
			wantErrIs:       ErrMissingRightSideExpr,
			wantErrContains: "logical operator without a right side expr in: \"name=alice or\"",
		},
		{
			name:            "err-unexpected-token",
			raw:             "name==eve",
			wantErrIs:       ErrUnexpectedToken,
			wantErrContains: `unexpected token "=" in: "name==eve"`,
		},
		{
			name:            "err-unexpected-logical-op",
			raw:             "name=alice and and description=friend",
			wantErrIs:       ErrUnexpectedLogicalOp,
			wantErrContains: `unexpected logical operator "and" when we've already parsed one for expr in: "name=alice and and description=friend"`,
		},
		{
			name:            "err-missing-logical-op",
			raw:             "name=alice description=friend",
			wantErrIs:       ErrUnexpectedExpr,
			wantErrContains: `unexpected expression starting at "description" in: "name=alice description=friend"`,
		},
		{
			name:            "err-unexpected-closing-paren",
			raw:             ")(name=alice)",
			wantErrIs:       ErrUnexpectedClosingParen,
			wantErrContains: `unexpected closing paren ")" but we haven't parsed a left side expression in: ")(name=alice)"`,
		},
		{
			name:            "err-unexpected-opening-paren",
			raw:             "((name=alice)",
			wantErrIs:       ErrMissingClosingParen,
			wantErrContains: `missing closing paren in: "((name=alice)"`,
		},
		{
			name:            "err-invalid-not-equal-after-whitespace",
			raw:             "   !not",
			wantErrIs:       ErrInvalidNotEqual,
			wantErrContains: `mql.lexNotEqualState: invalid "!=" token, got "!n"`,
		},
		{
			name: "success-double-parens",
			raw:  "((name=alice))",
			want: &comparisonExpr{
				column:       "name",
				comparisonOp: "=",
				value:        pointer("alice"),
			},
		},
		{
			name: "success-logical-expr-with-contains",
			raw:  "name=alice and address%\"my town\"",
			want: &logicalExpr{
				leftExpr: &comparisonExpr{
					column:       "name",
					comparisonOp: "=",
					value:        pointer("alice"),
				},
				logicalOp: "and",
				rightExpr: &comparisonExpr{
					column:       "address",
					comparisonOp: "%",
					value:        pointer("my town"),
				},
			},
		},
		{
			name: "nested-logical-expr",
			raw:  "(name=alice and address%hometown) or age > 21.5",
			want: &logicalExpr{
				leftExpr: &logicalExpr{
					leftExpr: &comparisonExpr{
						column:       "name",
						comparisonOp: "=",
						value:        pointer("alice"),
					},
					logicalOp: "and",
					rightExpr: &comparisonExpr{
						column:       "address",
						comparisonOp: "%",
						value:        pointer("hometown"),
					},
				},
				logicalOp: "or",
				rightExpr: &comparisonExpr{
					column:       "age",
					comparisonOp: ">",
					value:        pointer("21.5"),
				},
			},
		},
		{
			name: "reverse-nested-logical-expr",
			raw:  "age > 21.5 or (name=alice and address%hometown)",
			want: &logicalExpr{
				leftExpr: &comparisonExpr{
					column:       "age",
					comparisonOp: ">",
					value:        pointer("21.5"),
				},
				logicalOp: "or",
				rightExpr: &logicalExpr{
					leftExpr: &comparisonExpr{
						column:       "name",
						comparisonOp: "=",
						value:        pointer("alice"),
					},
					logicalOp: "and",
					rightExpr: &comparisonExpr{
						column:       "address",
						comparisonOp: "%",
						value:        pointer("hometown"),
					},
				},
			},
		},
		{
			name: "reverse-nested-logical-expr",
			raw:  `name=one or (created_at>"now()-interval '1 day'")`,
			want: &logicalExpr{
				leftExpr: &comparisonExpr{
					column:       "name",
					comparisonOp: "=",
					value:        pointer("one"),
				},
				logicalOp: "or",
				rightExpr: &comparisonExpr{
					column:       "created_at",
					comparisonOp: ">",
					value:        pointer("now()-interval '1 day'"),
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			p := newParser(tc.raw)
			e, err := p.parse()
			if tc.wantErrContains != "" {
				require.Errorf(err, "expected err for %s, but got %v", tc.raw, e)
				assert.Empty(e)
				if tc.wantErrIs != nil {
					assert.ErrorIs(err, tc.wantErrIs)
				}
				assert.ErrorContains(err, tc.wantErrContains)
				return
			}
			require.NoErrorf(err, "unexpected err for %s, but got %v", tc.raw, e)
			assert.Equal(tc.want, e)
		})
	}
}

// Fuzz_parserParse is primarily focused on finding panics
func Fuzz_parserParse(f *testing.F) {
	tc := []string{">=!=", "name=default OR age", "< <= = != AND OR and or", "1  !=   \"2\"", "(Name=\"Alice Eve\")"}
	for _, tc := range tc {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, s string) {
		parser := newParser(s)
		expr, err := parser.parse()
		if err == nil {
			switch expr.Type() {
			case logicalExprType, comparisonExprType:
			default:
				t.Errorf("unexpected expr: %v", expr)
			}
		}
	})
}

func Test_scan(t *testing.T) {
	t.Parallel()
	// just negative tests
	t.Run("err-scan-options", func(t *testing.T) {
		p := newParser("name=alice")
		err := p.scan(WithConverter("name", nil))
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidParameter)
		assert.ErrorContains(t, err, "missing ConvertToSqlFunc: invalid parameter")
	})
}
