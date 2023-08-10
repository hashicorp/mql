package mql_test

import (
	"database/sql"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/hashicorp/mql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testModel struct {
	ID           uint
	Name         string
	Email        *string
	Age          uint8
	Length       float32
	Birthday     *time.Time
	MemberNumber sql.NullString
	ActivatedAt  sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		query           string
		model           any
		opts            []mql.Option
		want            *mql.WhereClause
		wantErrContains string
		wantErrIs       error
	}{
		{
			name:  "success",
			query: "(name=alice and email=eve@example.com and member_number = 1) or (age > 21 or length < 1.5)",
			model: &testModel{},
			want: &mql.WhereClause{
				Condition: "(((name=? and email=?) and member_number=?) or (age>? or length<?))",
				Args:      []any{"alice", "eve@example.com", "1", 21, 1.5},
			},
		},
		{
			name:  "null-string",
			query: "name=null",
			model: &testModel{},
			want: &mql.WhereClause{
				Condition: "name=?",
				Args:      []any{"null"},
			},
		},
		{
			name:  "success-contains",
			query: "name%alice",
			model: testModel{},
			want: &mql.WhereClause{
				Condition: "name like ?",
				Args:      []any{"%alice%"},
			},
		},
		{
			name:            "err-leftExpr-without-op",
			query:           "age (name=alice)",
			model:           testModel{},
			wantErrIs:       mql.ErrUnexpectedOpeningParen,
			wantErrContains: `unexpected opening paren in: "age (name=alice)"`,
		},
		{
			name:            "err-int-model",
			query:           "name=alice",
			model:           1,
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "model must be a struct or a pointer to a struct",
		},
		{
			name:            "err-*int-model",
			query:           "name=alice",
			model:           pointer(1),
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "model must be a struct or a pointer to a struct",
		},
		{
			name:  "err-interface-nil-pointer-model",
			query: "name=alice",
			model: func() any {
				var r io.Reader
				return r
			}(),
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "missing model: invalid parameter",
		},
		{
			name:  "time",
			query: "created_at=\"2023-01-02\"",
			model: testModel{},
			want: &mql.WhereClause{
				Condition: "created_at::date=?",
				Args:      []any{"2023-01-02"},
			},
		},
		{
			name:  "success-with-column-map",
			query: "custom_name=alice",
			model: testModel{},
			opts:  []mql.Option{mql.WithColumnMap(map[string]string{"custom_name": "name"})},
			want: &mql.WhereClause{
				Condition: "name=?",
				Args:      []any{"alice"},
			},
		},
		{
			name:  "err-WithConverter-missing-field-name",
			query: "name=alice",
			model: testModel{},
			opts: []mql.Option{
				mql.WithConverter(
					"",
					func(columnName string, comparisonOp mql.ComparisonOp, value *string) (*mql.WhereClause, error) {
						return &mql.WhereClause{Condition: "name=?", Args: []any{"alice"}}, nil
					},
				),
			},
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "missing field name: invalid parameter",
		},
		{
			name:  "success-WithConverter",
			query: "(name = alice and email=eve@example.com) or age > 21",
			model: testModel{},
			opts: []mql.Option{
				mql.WithConverter(
					"name",
					func(columnName string, comparisonOp mql.ComparisonOp, value *string) (*mql.WhereClause, error) {
						return &mql.WhereClause{
							// intentionally not the correct condition and
							// args, but this makes verifying the test
							// easier.
							Condition: fmt.Sprintf("success-WithConverter: %s%s?", columnName, comparisonOp),
							Args:      []any{"success-WithConverter: alice"},
						}, nil
					},
				),
			},
			want: &mql.WhereClause{
				Condition: "((success-WithConverter: name=? and email=?) or age>?)",
				Args:      []any{"success-WithConverter: alice", "eve@example.com", 21},
			},
		},
		{
			name:            "err-ignored-field-used-in-query",
			query:           "email=eve@example.com or name=alice",
			model:           &testModel{},
			opts:            []mql.Option{mql.WithIgnoredFields("Name")},
			wantErrContains: `mql.exprToWhereClause: invalid right expr: mql.exprToWhereClause: invalid column "name"`,
		},
		{
			name:            "err-missing-query",
			query:           "",
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "missing query: invalid parameter",
		},
		{
			name:            "err-model",
			query:           "name=alice",
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "missing model: invalid parameter",
		},
		{
			name:            "err-invalid-query",
			query:           "name!alice",
			model:           testModel{},
			wantErrIs:       mql.ErrInvalidNotEqual,
			wantErrContains: `invalid "!=" token, got "!a"`,
		},
		{
			name:            "err-invalid-WithConverter-opt",
			query:           "name=alice",
			model:           testModel{},
			opts:            []mql.Option{mql.WithConverter("TestColumn", nil)},
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "missing ConvertToSqlFunc: invalid parameter",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			whereClause, err := mql.Parse(tc.query, tc.model, tc.opts...)
			if tc.wantErrContains != "" {
				require.Errorf(err, "expected err for %s, but got %v", tc.query, whereClause)
				assert.Empty(whereClause)
				if tc.wantErrIs != nil {
					assert.ErrorIs(err, tc.wantErrIs)
				}
				assert.ErrorContains(err, tc.wantErrContains)
				return
			}
			require.NoErrorf(err, "unexpected err for %s, but got %v", tc.query, whereClause)
			assert.Equal(tc.want, whereClause)
		})
	}
}

func pointer[T any](input T) *T {
	return &input
}
