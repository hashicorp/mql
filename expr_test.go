// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_root will focus on error conditions
func Test_root(t *testing.T) {
	t.Parallel()
	t.Run("missing-expr", func(t *testing.T) {
		e, err := root(nil, "raw")
		require.Error(t, err)
		assert.Empty(t, e)
		assert.ErrorIs(t, err, ErrInvalidParameter)
		assert.ErrorContains(t, err, "invalid parameter (missing expression)")
	})
	t.Run("missing-left-expr", func(t *testing.T) {
		e, err := root(&logicalExpr{
			leftExpr:  nil,
			logicalOp: "",
			rightExpr: &comparisonExpr{},
		}, "raw")
		require.Error(t, err)
		assert.Empty(t, e)
		assert.ErrorIs(t, err, ErrMissingExpr)
		assert.ErrorContains(t, err, "missing expression nil in: \"raw\"")
	})
}

// Test_newComparison will focus on error conditions
func Test_newLogicalOp(t *testing.T) {
	t.Parallel()
	t.Run("invalid-comp-op", func(t *testing.T) {
		op, err := newLogicalOp("not-valid")
		require.Error(t, err)
		assert.Empty(t, op)
		assert.ErrorIs(t, err, ErrInvalidLogicalOp)
		assert.ErrorContains(t, err, `invalid logical operator "not-valid"`)
	})
}

// Test_newComparisonOp will focus on error conditions
func Test_newComparisonOp(t *testing.T) {
	t.Parallel()
	t.Run("invalid-comp-op", func(t *testing.T) {
		op, err := newComparisonOp("not-valid")
		require.Error(t, err)
		assert.Empty(t, op)
		assert.ErrorIs(t, err, ErrInvalidComparisonOp)
		assert.ErrorContains(t, err, `invalid comparison operator "not-valid"`)
	})
}

func Test_comparisonExprString(t *testing.T) {
	t.Run("nil-value", func(t *testing.T) {
		e := &comparisonExpr{
			column:       "name",
			comparisonOp: "=",
			value:        nil,
		}
		assert.Equal(t, "(comparisonExpr: name = nil)", e.String())
	})
}

func Test_logicalExprString(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		e := &logicalExpr{
			leftExpr: &comparisonExpr{
				column:       "name",
				comparisonOp: "=",
				value:        pointer("alice"),
			},
			logicalOp: andOp,
			rightExpr: &comparisonExpr{
				column:       "name",
				comparisonOp: "=",
				value:        pointer("alice"),
			},
		}
		assert.Equal(t, "(logicalExpr: (comparisonExpr: name = alice) and (comparisonExpr: name = alice))", e.String())
	})
}

// Test_defaultValidateConvert will focus on error conditions
func Test_defaultValidateConvert(t *testing.T) {
	t.Parallel()
	fValidators, err := fieldValidators(reflect.ValueOf(testModel{}))
	require.NoError(t, err)
	opts := getDefaultOptions()
	t.Run("missing-column", func(t *testing.T) {
		e, err := defaultValidateConvert("", EqualOp, pointer("alice"), fValidators["name"], opts)
		require.Error(t, err)
		assert.Empty(t, e)
		assert.ErrorIs(t, err, ErrMissingColumn)
		assert.ErrorContains(t, err, "missing column")
	})
	t.Run("missing-comparison-op", func(t *testing.T) {
		e, err := defaultValidateConvert("name", "", pointer("alice"), fValidators["name"], opts)
		require.Error(t, err)
		assert.Empty(t, e)
		assert.ErrorIs(t, err, ErrMissingComparisonOp)
		assert.ErrorContains(t, err, "missing comparison operator")
	})
	t.Run("missing-value", func(t *testing.T) {
		e, err := defaultValidateConvert("name", EqualOp, nil, fValidators["name"], opts)
		require.Error(t, err)
		assert.Empty(t, e)
		assert.ErrorIs(t, err, ErrMissingComparisonValue)
		assert.ErrorContains(t, err, "missing comparison value")
	})
	t.Run("missing-validator-func", func(t *testing.T) {
		e, err := defaultValidateConvert("name", EqualOp, pointer("alice"), validator{typ: "string"}, opts)
		require.Error(t, err)
		assert.Empty(t, e)
		assert.ErrorIs(t, err, ErrInvalidParameter)
		assert.ErrorContains(t, err, "missing validator function")
	})
	t.Run("missing-validator-typ", func(t *testing.T) {
		e, err := defaultValidateConvert("name", EqualOp, pointer("alice"), validator{fn: fValidators["name"].fn}, opts)
		require.Error(t, err)
		assert.Empty(t, e)
		assert.ErrorIs(t, err, ErrInvalidParameter)
		assert.ErrorContains(t, err, "missing validator type")
	})
	t.Run("success-with-table-override", func(t *testing.T) {
		opts.withTableColumnMap["name"] = "users.name"
		e, err := defaultValidateConvert("name", EqualOp, pointer("alice"), validator{fn: fValidators["name"].fn, typ: "default"}, opts)
		assert.Empty(t, err)
		assert.NotEmpty(t, e)
		assert.Equal(t, "users.name=?", e.Condition, "condition")
		assert.Len(t, e.Args, 1, "args")
		assert.Equal(t, "alice", e.Args[0], "args[0]")
	})
}
