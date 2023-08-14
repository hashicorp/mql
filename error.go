// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql

import "errors"

var (
	ErrInternal               = errors.New("internal error")
	ErrInvalidParameter       = errors.New("invalid parameter")
	ErrInvalidNotEqual        = errors.New(`invalid "!=" token`)
	ErrMissingExpr            = errors.New("missing expression")
	ErrUnexpectedExpr         = errors.New("unexpected expression")
	ErrUnexpectedClosingParen = errors.New("unexpected closing paren")
	ErrMissingClosingParen    = errors.New("missing closing paren")
	ErrUnexpectedOpeningParen = errors.New("unexpected opening paren")
	ErrUnexpectedLogicalOp    = errors.New("unexpected logical operator")
	ErrUnexpectedToken        = errors.New("unexpected token")
	ErrInvalidComparisonOp    = errors.New("invalid comparison operator")
	ErrMissingComparisonOp    = errors.New("missing comparison operator")
	ErrMissingColumn          = errors.New("missing column")
	ErrInvalidLogicalOp       = errors.New("invalid logical operator")
	ErrMissingLogicalOp       = errors.New("missing logical operator")
	ErrMissingRightSideExpr   = errors.New("logical operator without a right side expr")
	ErrMissingComparisonValue = errors.New("missing comparison value")
	ErrInvalidColumn          = errors.New("invalid column")
)
