// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql

import (
	"fmt"
)

type options struct {
	withSkipWhitespace        bool
	withColumnMap             map[string]string
	withValidateConvertFn     ValidateConvertFunc
	withValidateConvertColumn string
	withIgnoredFields         []string
	withPgPlaceholder         bool
}

// Option - how options are passed as args
type Option func(*options) error

func getDefaultOptions() options {
	return options{
		withColumnMap: make(map[string]string),
	}
}

func getOpts(opt ...Option) (options, error) {
	opts := getDefaultOptions()

	for _, o := range opt {
		if err := o(&opts); err != nil {
			return opts, err
		}
	}
	return opts, nil
}

// withSkipWhitespace provides an option to request that whitespace be skipped
func withSkipWhitespace() Option {
	return func(o *options) error {
		o.withSkipWhitespace = true
		return nil
	}
}

// WithColumnMap provides an optional map of columns from a column in the user
// provided query to a column in the database model
func WithColumnMap(m map[string]string) Option {
	return func(o *options) error {
		if !isNil(m) {
			o.withColumnMap = m
		}
		return nil
	}
}

// ValidateConvertFunc validates the value and then converts the columnName,
// comparisonOp and value to a WhereClause
type ValidateConvertFunc func(columnName string, comparisonOp ComparisonOp, value *string) (*WhereClause, error)

// WithConverter provides an optional ConvertFunc for a column identifier in the
// query. This allows you to provide whatever custom validation+conversion you
// need on a per column basis.  See: DefaultValidateConvert(...) for inspiration.
func WithConverter(fieldName string, fn ValidateConvertFunc) Option {
	const op = "mql.WithSqlConverter"
	return func(o *options) error {
		switch {
		case fieldName != "" && !isNil(fn):
			o.withValidateConvertFn = fn
			o.withValidateConvertColumn = fieldName
		case fieldName == "" && !isNil(fn):
			return fmt.Errorf("%s: missing field name: %w", op, ErrInvalidParameter)
		case fieldName != "" && isNil(fn):
			return fmt.Errorf("%s: missing ConvertToSqlFunc: %w", op, ErrInvalidParameter)
		}
		return nil
	}
}

// WithIgnoredFields provides an optional list of fields to ignore in the model
// (your Go struct) when parsing. Note: Field names are case sensitive.
func WithIgnoredFields(fieldName ...string) Option {
	return func(o *options) error {
		if len(fieldName) > 0 {
			o.withIgnoredFields = make([]string, len(fieldName))
			for _, name := range fieldName {
				o.withIgnoredFields = append(o.withIgnoredFields, name)
			}
			o.withIgnoredFields = fieldName
		}
		return nil
	}
}

// WithPgPlaceholders will use parameters placeholders that are compatible with
// the postgres pg driver which requires a placeholder like $1 instead of ?.
// See:
//   - https://pkg.go.dev/github.com/jackc/pgx/v5
//   - https://pkg.go.dev/github.com/lib/pq
func WithPgPlaceholders() Option {
	return func(o *options) error {
		o.withPgPlaceholder = true
		return nil
	}
}
