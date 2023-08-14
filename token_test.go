// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_lexerString(t *testing.T) {
	for typ, s := range tokenTypeToString {
		assert := assert.New(t)
		assert.Equal(s, typ.String())
	}
	t.Run("unknown-tokenType", func(t *testing.T) {
		typ := tokenType(-1)
		assert.Equal(t, tokenTypeToString[unknownToken], typ.String())
	})
}
