// Copyright (c) HashiCorp, Inc.

package mql

import (
	"fmt"
)

type exprType int

const (
	comparisonExprType exprType = iota
	logicalExprType
)

type expr interface {
	Type() exprType
	String() string
}

type comparisonOp string

const (
	greaterThanOp        comparisonOp = ">"
	greaterThanOrEqualOp              = ">="
	lessThanOp                        = "<"
	lessThanOrEqualOp                 = "<="
	equalOp                           = "="
	notEqualOp                        = "!="
	containsOp                        = "%"
)

func newComparisonOp(s string) (comparisonOp, error) {
	const op = "newComparisonOp"
	switch s {
	case
		string(greaterThanOp),
		string(greaterThanOrEqualOp),
		string(lessThanOp),
		string(lessThanOrEqualOp),
		string(equalOp),
		string(notEqualOp),
		string(containsOp):
		return comparisonOp(s), nil
	default:
		return "", fmt.Errorf("%s: %w %q", op, ErrInvalidComparisonOp, s)
	}
}

type comparisonExpr struct {
	column       string
	comparisonOp comparisonOp
	value        *string
}

// Type returns the expr type
func (e *comparisonExpr) Type() exprType {
	return comparisonExprType
}

// String returns a string rep of the expr
func (e *comparisonExpr) String() string {
	switch e.value {
	case nil:
		return fmt.Sprintf("(comparisonExpr: %s %s nil)", e.column, e.comparisonOp)
	default:
		return fmt.Sprintf("(comparisonExpr: %s %s %s)", e.column, e.comparisonOp, *e.value)
	}
}

func (e *comparisonExpr) isComplete() bool {
	return e.column != "" && e.comparisonOp != "" && e.value != nil
}

type logicalOp string

const (
	andOp logicalOp = "and"
	orOp            = "or"
)

func newLogicalOp(s string) (logicalOp, error) {
	const op = "newLogicalOp"
	switch s {
	case
		string(andOp),
		string(orOp):
		return logicalOp(s), nil
	default:
		return "", fmt.Errorf("%s: %w %q", op, ErrInvalidLogicalOp, s)
	}
}

type logicalExpr struct {
	leftExpr  expr
	logicalOp logicalOp
	rightExpr expr
}

// Type returns the expr type
func (l *logicalExpr) Type() exprType {
	return logicalExprType
}

// String returns a string rep of the expr
func (l *logicalExpr) String() string {
	return fmt.Sprintf("(logicalExpr: %s %s %s)", l.leftExpr, l.logicalOp, l.rightExpr)
}

// root will return the root of the expr tree
func root(lExpr *logicalExpr, raw string) (expr, error) {
	const op = "mql.root"
	switch {
	// intentionally not checking raw, since can be an empty string
	case lExpr == nil:
		return nil, fmt.Errorf("%s: %w (missing expression)", op, ErrInvalidParameter)
	}
	logicalOp := lExpr.logicalOp
	if logicalOp != "" && lExpr.rightExpr == nil {
		return nil, fmt.Errorf("%s: %w in: %q", op, ErrMissingRightSideExpr, raw)
	}

	for lExpr.logicalOp == "" {
		switch {
		case lExpr.leftExpr == nil:
			return nil, fmt.Errorf("%s: %w nil in: %q", op, ErrMissingExpr, raw)
		case lExpr.leftExpr.Type() == comparisonExprType:
			return lExpr.leftExpr, nil
		default:
			lExpr = lExpr.leftExpr.(*logicalExpr)
		}
	}
	return lExpr, nil
}
