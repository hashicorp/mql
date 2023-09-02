// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

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

func Test_exprToWhereClause(t *testing.T) {
	t.Parallel()
	testValidators, err := fieldValidators(reflect.ValueOf(testModel{}))
	require.NoError(t, err)

	testParser := newParser("name=alice")
	testExpr, err := testParser.parse()
	require.NoError(t, err)

	tests := []struct {
		name            string
		expr            expr
		validators      map[string]validator
		opt             []Option
		want            *WhereClause
		wantErrContains string
		wantErrIs       error
	}{
		{
			name:            "err-missing-expr",
			validators:      testValidators,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "missing expression: invalid parameter",
		},
		{
			name:            "err-missing-validators",
			expr:            testExpr,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "missing validators: invalid parameter",
		},
		{
			name:            "err-invalid-converter",
			expr:            testExpr,
			validators:      testValidators,
			opt:             []Option{WithConverter("name", nil)},
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "missing ConvertToSqlFunc: invalid parameter",
		},
		{
			name: "invalid-float",
			expr: &comparisonExpr{
				column:       "length",
				comparisonOp: "=",
				value:        pointer("1.11."),
			},
			validators:      testValidators,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: `"1.11." in (comparisonExpr: length = 1.11.): invalid parameter`,
		},
		{
			name: "invalid-int",
			expr: &comparisonExpr{
				column:       "age",
				comparisonOp: "=",
				value:        pointer("1.11"),
			},
			validators:      testValidators,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: `"1.11" in (comparisonExpr: age = 1.11): invalid parameter`,
		},
		{
			name: "err-invalid-logicalExpr-left",
			expr: &logicalExpr{
				leftExpr: &comparisonExpr{
					column:       "name",
					comparisonOp: "",
					value:        nil,
				},
				logicalOp: "and",
				rightExpr: &comparisonExpr{
					column:       "name",
					comparisonOp: "=",
					value:        pointer("alice"),
				},
			},
			validators:      testValidators,
			wantErrContains: "invalid left expr: mql.exprToWhereClause: mql.(comparisonExpr).convertToSql: missing comparison operator",
		},
		{
			name: "err-missing-logicalOp",
			expr: &logicalExpr{
				leftExpr:  testExpr,
				rightExpr: testExpr,
			},
			validators:      testValidators,
			wantErrIs:       ErrMissingLogicalOp,
			wantErrContains: `missing logical operator that stated with left expr condition: "name=?" args: ["alice"]`,
		},
		{
			name:            "err-unexpected-expr-type",
			expr:            &invalidExpr{},
			validators:      testValidators,
			wantErrIs:       ErrInternal,
			wantErrContains: `unexpected expr type *mql.invalidExpr: internal error`,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			whereClause, err := exprToWhereClause(tc.expr, tc.validators, tc.opt...)
			if tc.wantErrContains != "" {
				require.Errorf(err, "expected err for %s but got %v", tc.expr, whereClause)
				assert.Empty(whereClause)
				if tc.wantErrIs != nil {
					assert.ErrorIs(err, tc.wantErrIs)
				}
				assert.ErrorContains(err, tc.wantErrContains)
				return
			}
			require.NoErrorf(err, "no error expected for %s but got: %s", tc.expr, err)
			assert.Equal(tc.want, whereClause)
		})
	}
}

func Test_fieldValidators(t *testing.T) {
	// just a few negative tests
	t.Parallel()
	t.Run("zero-value-model", func(t *testing.T) {
		_, err := fieldValidators(reflect.Value{})
		require.Error(t, err)
		assert.ErrorContains(t, err, "missing model: invalid parameter")
		assert.ErrorIs(t, err, ErrInvalidParameter)
	})
}

type invalidExpr struct{}

func (_ *invalidExpr) Type() exprType {
	return unknownExprType
}
func (_ *invalidExpr) String() string {
	return "unknown"
}
