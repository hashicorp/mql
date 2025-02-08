// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mql_test

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
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
			query: "(name=\"alice\" and email=\"eve@example.com\" and member_number = 1) or (age > 21 or length < 1.5)",
			model: &testModel{},
			want: &mql.WhereClause{
				Condition: "(((name=? and email=?) and member_number=?) or (age>? or length<?))",
				Args:      []any{"alice", "eve@example.com", "1", 21, 1.5},
			},
		},
		{
			name:  "success-single-quote-delimiters",
			query: "(name='alice' and email='eve@example.com' and member_number = 1) or (age > 21 or length < 1.5)",
			model: &testModel{},
			want: &mql.WhereClause{
				Condition: "(((name=? and email=?) and member_number=?) or (age>? or length<?))",
				Args:      []any{"alice", "eve@example.com", "1", 21, 1.5},
			},
		},
		{
			name:  "success-backtick-delimiters",
			query: "(name=`alice` and email=`eve@example.com` and member_number = 1) or (age > 21 or length < 1.5)",
			model: &testModel{},
			want: &mql.WhereClause{
				Condition: "(((name=? and email=?) and member_number=?) or (age>? or length<?))",
				Args:      []any{"alice", "eve@example.com", "1", 21, 1.5},
			},
		},
		{
			name:  "success-multi-columned",
			query: "(name=`alice`) and (email=`eve@example.com`) and (member_number = 1)",
			model: &testModel{},
			want: &mql.WhereClause{
				Condition: "(name=? and (email=? and member_number=?))",
				Args:      []any{"alice", "eve@example.com", "1"},
			},
		},
		{
			name:  "success-multi-columned-with-an-or",
			query: "(name=`alice`) and (email=`eve@example.com`) or (member_number = 1)",
			model: &testModel{},
			want: &mql.WhereClause{
				Condition: "(name=? and (email=? or member_number=?))",
				Args:      []any{"alice", "eve@example.com", "1"},
			},
		},
		{
			name:  "null-string",
			query: "name=\"null\"",
			model: &testModel{},
			want: &mql.WhereClause{
				Condition: "name=?",
				Args:      []any{"null"},
			},
		},
		{
			name:  "success-contains",
			query: "name%\"alice\"",
			model: testModel{},
			want: &mql.WhereClause{
				Condition: "name like ?",
				Args:      []any{"%alice%"},
			},
		},
		{
			name:  "success-WithPgPlaceholder",
			query: "name=\"bob\" or (name%\"alice\" or name=\"eve\")",
			model: testModel{},
			opts:  []mql.Option{mql.WithPgPlaceholders()},
			want: &mql.WhereClause{
				Condition: "(name=$1 or (name like $2 or name=$3))",
				Args:      []any{"bob", "%alice%", "eve"},
			},
		},
		{
			name:  "success-dd",
			query: "nAme%\"\"",
			model: &testModel{},
			want: &mql.WhereClause{
				Condition: "name like ?",
				Args:      []any{"%%"},
			},
		},
		{
			name:            "err-leftExpr-without-op",
			query:           "age (name=\"alice\")",
			model:           testModel{},
			wantErrIs:       mql.ErrUnexpectedOpeningParen,
			wantErrContains: `unexpected opening paren in: "age (name=\"alice\")"`,
		},
		{
			name:            "err-int-model",
			query:           "name=\"alice\"",
			model:           1,
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "model must be a struct or a pointer to a struct",
		},
		{
			name:            "err-*int-model",
			query:           "name=\"alice\"",
			model:           pointer(1),
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "model must be a struct or a pointer to a struct",
		},
		{
			name:  "err-interface-nil-pointer-model",
			query: "name=\"alice\"",
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
			query: "custom_name=\"alice\"",
			model: testModel{},
			opts:  []mql.Option{mql.WithColumnMap(map[string]string{"custom_name": "name"})},
			want: &mql.WhereClause{
				Condition: "name=?",
				Args:      []any{"alice"},
			},
		},
		{
			name:  "err-WithConverter-missing-field-name",
			query: "name=\"alice\"",
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
			name:  "err-WithConverter-duplicated-converter",
			query: "name=\"alice\" and email=\"eve@example.com\"",
			model: testModel{},
			opts: []mql.Option{
				mql.WithConverter(
					"name",
					func(columnName string, comparisonOp mql.ComparisonOp, value *string) (*mql.WhereClause, error) {
						return &mql.WhereClause{Condition: "name=?", Args: []any{"alice"}}, nil
					},
				),
				mql.WithConverter(
					"email",
					func(columnName string, comparisonOp mql.ComparisonOp, value *string) (*mql.WhereClause, error) {
						return &mql.WhereClause{Condition: "email=?", Args: []any{"eve@example.com"}}, nil
					},
				),
				mql.WithConverter(
					"name",
					func(columnName string, comparisonOp mql.ComparisonOp, value *string) (*mql.WhereClause, error) {
						return &mql.WhereClause{Condition: "duplicated-Converter name=?", Args: []any{"alice"}}, nil
					},
				),
			},
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "duplicated convert: invalid parameter",
		},
		{
			name:  "success-WithConverter",
			query: "(name = \"alice\" and email=\"eve@example.com\") or age > 21",
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
			name:  "success-WithMultiConverters",
			query: "(name = \"alice\" and email=\"eve@example.com\") or age > 21",
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
				mql.WithConverter(
					"email",
					func(columnName string, comparisonOp mql.ComparisonOp, value *string) (*mql.WhereClause, error) {
						return &mql.WhereClause{
							// intentionally not the correct condition and
							// args, but this makes verifying the test
							// easier.
							Condition: fmt.Sprintf("success-WithConverter: %s%s?", columnName, comparisonOp),
							Args:      []any{"success-WithConverter: email=\"eva@example.com\""},
						}, nil
					},
				),
			},
			want: &mql.WhereClause{
				Condition: "((success-WithConverter: name=? and success-WithConverter: email=?) or age>?)",
				Args:      []any{"success-WithConverter: alice", "success-WithConverter: email=\"eva@example.com\"", 21},
			},
		},
		{
			name:            "err-ignored-field-used-in-query",
			query:           "email=\"eve@example.com\" or name=\"alice\"",
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
			query:           "name=\"alice\"",
			model:           testModel{},
			opts:            []mql.Option{mql.WithConverter("TestColumn", nil)},
			wantErrIs:       mql.ErrInvalidParameter,
			wantErrContains: "missing ConvertToSqlFunc: invalid parameter",
		},
		{
			name:  "success-with-table-column-map",
			query: "custom_name=\"alice\"",
			model: testModel{},
			opts: []mql.Option{
				mql.WithColumnMap(map[string]string{"custom_name": "name"}),
				mql.WithTableColumnMap(map[string]string{"name": "users.custom->>'name'"}),
			},
			want: &mql.WhereClause{
				Condition: "users.custom->>'name'=?",
				Args:      []any{"alice"},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
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

// Fuzz_mqlParse is primarily focused on finding sql injection and panics
func Fuzz_mqlParse(f *testing.F) {
	tc := []string{
		">=!=",
		"name=default OR age",
		"< <= = != AND OR and or",
		"1  !=   \"2\"",
		"(Name=\"Alice Eve\")",
		`name="alice"`,
		`name="alice\\eve"`,
		`name='alice'`,
		"name=`alice's`",
	}
	for _, tc := range tc {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, s string) {
		where, err := mql.Parse(s, testModel{})
		if err == nil {
			for _, kw := range sqlKeywordsExceptLike {
				if strings.Contains(strings.ToLower(where.Condition), kw) {
					t.Errorf("unexpected sql keyword %q in %s", kw, where.Condition)
				}
			}
		}
	})
}

var sqlKeywordsExceptLike = []string{
	"select", "from", "where", "join", "left", "right", "inner", "outer",
	"on", "group", "by", "order", "having", "insert", "update", "delete",
	"values", "set", "as", "distinct", "limit", "offset", "and", "or",
	"not", "in", "between", "is", "null", "true", "false",
	"case", "when", "then", "else", "end", "while", "for", "foreach",
	"create", "alter", "drop", "table", "view", "index", "sequence",
	"database", "schema", "function", "procedure", "trigger", "event",
	"primary", "foreign", "references", "constraint", "unique",
	"default", "auto_increment", "check", "cascade",
	"explain", "analyze", "describe",
	"primary", "foreign", "key", "index", "references", "check", "tablespace",
	"sequence", "constraint", "default", "charset", "collate", "column",
	"table", "view", "materialized", "index", "trigger", "domain",
	"data", "type", "array", "enum", "oid", "range", "returning",
	"inherits", "rule", "with", "time", "zone", "at", "serializable",
	"repeatable", "committed", "uncommitted", "isolation", "lock",
	"share", "mode", "nowait", "wait", "array_agg", "avg", "count",
	"max", "min", "cast", "convert", "overlaps", "date",
	"time", "timestamp", "extract", "current_date", "current_time",
	"current_timestamp", "now", "current_user", "current_schema",
	"transaction", "true", "false", "unknown", "absolute", "relative",
	"forward", "backward", "transaction", "level", "read", "immediate",
	"deferred", "none", "autocommit", "off", "on", "savepoint",
	"rollback", "release", "chain", "cascaded", "local", "session",
	"global", "temporary", "temp", "unsigned", "signed", "precision",
	"first", "next", "both", "prior", "absolute", "relative", "forward",
	"backward", "localtime", "localtimestamp", "timeofday",
	"array", "row", "multiset", "map", "json", "xml", "struct", "clob",
	"blob", "nclob", "bytea", "jsonb", "jsonpath", "xmltype", "tinyint",
	"smallint", "integer", "bigint", "decimal", "numeric", "real",
	"double", "float", "character", "char", "varchar", "nchar",
	"nvarchar", "binary", "varbinary", "timestamp", "interval",
	"year", "month", "day", "hour", "minute", "second", "zone",
	"boolean", "bit", "enum", "set", "uuid", "oid", "cidr", "inet",
	"macaddr", "serial", "bigserial", "money", "setof", "record",
	"anyelement", "anyarray", "anynonarray", "anyenum", "anyrange",
	"array_agg", "string_agg", "avg", "count", "max", "min",
	"sum", "stddev", "var_pop", "var_samp", "covar_pop",
	"covar_samp", "corr", "regr_avgx", "regr_avgy",
	"regr_count", "regr_intercept", "regr_r2", "regr_slope",
	"regr_sxx", "regr_sxy", "regr_syy", "bit_and", "bit_or",
	"bit_xor", "row_number", "rank", "dense_rank", "percent_rank",
	"cume_dist", "ntile", "first_value", "last_value", "lead",
	"lag", "percentile_cont", "percentile_disc", "mode", "with",
	"insensitive", "sensitive", "scroll", "cursor", "without",
	"type", "only", "precision", "double", "within",
	"zone", "over", "lead", "lag", "ignore",
	"nulls", "exclude", "ties", "from", "leading", "trailing",
	"both", "not", "first", "last", "after", "before", "each",
	"statement", "at", "at", "time", "zone", "serializable",
	"repeatable", "read", "committed", "uncommitted", "isolation",
	"level", "lock", "share", "mode", "nowait", "wait", "explain",
	"analyze", "describe", "cast", "convert", "to", "using",
	"explicit", "implicit", "inner", "cross", "left", "right",
	"outer", "full", "join", "using", "matched", "not", "then",
	"insert", "ignore", "into", "first", "last", "values", "null",
	"before", "after", "each", "row", "statement", "at", "time",
	"zone", "serializable", "repeatable", "read", "committed",
	"uncommitted", "isolation", "level", "on", "delete", "natural",
	"set", "default", "auto_increment", "check", "cascade", "with",
	"option", "modify", "auto_increment", "check", "cascade", "in",
	"out", "inout", "as", "insensitive", "sensitive", "language",
	"sql", "validator", "old", "new", "old_table", "new_table",
	"old_row", "new_row", "after_trigger", "before_trigger",
	"instead_of_trigger", "execute", "function", "procedure",
	"returns", "table", "return", "rows", "cursor", "inserting",
	"deleting", "updating", "after_statement", "before_statement",
	"declare", "condition", "signal", "resignal", "undo", "handler",
	"get", "diagnostics", "reset", "set", "position", "resume",
	"suspend", "leave", "iterate", "repeat", "until", "close",
	"fetch", "open", "prepare", "execute", "deallocate", "forward",
	"backward", "absolute", "relative", "release", "rollback",
	"work", "savepoint", "scroll", "replace", "escape", "glob",
	"regexp", "matches", "unknown", "cube", "rollup", "ordering",
	"search", "depth", "children", "siblings", "value", "positive",
	"negative", "union", "intersect", "except", "case", "cast",
	"convert", "current_date", "current_time", "current_timestamp",
	"date_part", "date_trunc", "extract", "localtime",
	"localtimestamp", "now", "timeofday", "timestampadd",
	"timestampdiff", "array_agg", "string_agg", "avg", "count",
	"max", "min", "sum", "stddev", "var_pop", "var_samp",
	"covar_pop", "covar_samp", "corr", "regr_avgx", "regr_avgy",
	"regr_count", "regr_intercept", "regr_r2", "regr_slope",
	"regr_sxx", "regr_sxy", "regr_syy", "bit_and", "bit_or",
	"bit_xor", "row_number", "rank", "dense_rank", "percent_rank",
	"cume_dist", "ntile", "first_value", "last_value", "lead",
	"lag", "percentile_cont", "percentile_disc", "mode", "with",
	"insensitive", "sensitive", "scroll", "cursor", "without",
	"type", "only", "first", "next", "both", "prior", "absolute",
	"relative", "forward", "backward", "transaction", "isolation",
	"level", "read", "uncommitted", "committed", "repeatable",
	"serializable", "immediate", "deferred", "explicit", "none",
	"current_schema", "current_user", "session_user", "system_user",
	"user", "autocommit", "off", "on", "savepoint", "rollback",
	"release", "work", "chain", "cascaded", "local", "release",
	"session", "global", "temporary", "temp", "unsigned", "signed",
	"precision", "double", "within", "zone", "over", "lead", "lag",
	"ignore", "nulls", "exclude", "ties", "from", "leading", "trailing",
	"both", "not", "first", "last", "after", "before", "each",
	"statement", "at", "time", "zone", "serializable", "repeatable",
	"read", "committed", "uncommitted", "isolation", "level",
	"lock", "share", "mode", "nowait", "wait", "explain", "analyze",
	"describe", "cast", "convert", "to", "using", "explicit",
	"implicit", "inner", "cross", "left", "right", "outer", "full",
	"join", "using", "matched", "not", "then", "insert", "ignore",
	"into", "first", "last", "values", "null", "before", "after",
	"each", "row", "statement", "at", "time", "zone", "serializable",
	"repeatable", "read", "committed", "uncommitted", "isolation",
	"level", "on", "delete", "natural", "set", "default",
	"auto_increment", "check", "cascade", "with", "option",
	"modify", "auto_increment", "check", "cascade", "in", "out",
	"inout", "as", "insensitive", "sensitive", "language", "sql",
	"validator", "old", "new", "old_table", "new_table", "old_row",
	"new_row", "after_trigger", "before_trigger",
	"instead_of_trigger", "execute", "function", "procedure",
	"returns", "table", "return", "rows", "cursor", "inserting",
	"deleting", "updating", "after_statement", "before_statement",
	"declare", "condition", "signal", "resignal", "undo", "handler",
	"get", "diagnostics", "reset", "set", "position", "resume",
	"suspend", "leave", "iterate", "repeat", "until", "close",
	"fetch", "open", "prepare", "execute", "deallocate", "forward",
	"backward", "absolute", "relative", "release", "rollback",
	"work", "savepoint", "scroll", "replace", "escape", "glob",
	"regexp", "matches", "unknown", "cube", "rollup", "ordering",
	"search", "depth", "children", "siblings", "value", "positive",
	"negative", "union", "intersect", "except", "case", "cast",
	"convert", "current_date", "current_time", "current_timestamp",
	"date_part", "date_trunc", "extract", "localtime",
	"localtimestamp", "now", "timeofday", "timestampadd",
	"timestampdiff", "array_agg", "string_agg", "avg", "count",
	"max", "min", "sum",
}
