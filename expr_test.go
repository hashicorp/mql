// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql

import (
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
