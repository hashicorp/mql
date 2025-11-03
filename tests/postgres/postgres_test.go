// Copyright IBM Corp. 2023, 2025
// SPDX-License-Identifier: MPL-2.0

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/mql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Test_postgres(t *testing.T) {
	t.Parallel()
	testCtx := context.Background()
	db := setupDB(t)
	rw := dbw.New(db)
	now := time.Now()
	testInsertUser(t, rw, &user{ID: 1, Name: "one", Email: pointer("one@example.com"), Age: 1, CreatedAt: now.Add(1 * 24 * time.Hour)})
	testInsertUser(t, rw, &user{ID: 2, Name: "two", Email: pointer("two@example.com"), Age: 2, CreatedAt: now.Add(2 * 24 * time.Hour)})
	tests := []struct {
		name            string
		query           string
		opts            []mql.Option
		wantErrContains string
		wantErrIs       error
		want            []*user
	}{
		{
			name:  "simple",
			query: `name="one" and age>0`,
			want:  []*user{{ID: 1, Name: "one", Email: pointer("one@example.com"), Age: 1, CreatedAt: now.Add(1 * 24 * time.Hour)}},
		},
		{
			name:  "WithConverter",
			query: fmt.Sprintf(`name="one" or (created_at>%q)`, time.Now().Add(2*24*time.Hour).Format("2006-01-02")),
			opts: []mql.Option{
				mql.WithConverter("created_at", func(columnName string, comparisonOp mql.ComparisonOp, value *string) (*mql.WhereClause, error) {
					return &mql.WhereClause{
						Condition: fmt.Sprintf("%s::date%s?", columnName, comparisonOp),
						Args:      []any{*value},
					}, nil
				}),
			},
			want: []*user{{ID: 1, Name: "one", Email: pointer("one@example.com"), Age: 1, CreatedAt: now.Add(1 * 24 * time.Hour)}},
		},
		{
			name:  "default-time-converter",
			query: fmt.Sprintf(`name="one" or (created_at>%q)`, time.Now().Add(2*24*time.Hour).Format("2006-01-02")),
			want:  []*user{{ID: 1, Name: "one", Email: pointer("one@example.com"), Age: 1, CreatedAt: now.Add(1 * 24 * time.Hour)}},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			where, err := mql.Parse(tc.query, user{}, tc.opts...)
			if tc.wantErrContains != "" {
				require.Error(err)
				assert.Empty(where)
				assert.ErrorContains(err, tc.wantErrContains)
				if tc.wantErrIs != nil {
					assert.ErrorIs(err, tc.wantErrIs)
				}
				return
			}
			require.NoError(err)
			{
				// test dbw
				var found []*user
				err = rw.SearchWhere(testCtx, &found, where.Condition, where.Args)
				require.NoError(err)
				datesWithinRange(t, tc.want, found)
				assert.Equal(tc.want, found)
			}
			{
				var found []*user
				sqlDB, err := db.SqlDB(testCtx)
				require.NoError(err)
				gormDB, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
					Conn: sqlDB,
				}), &gorm.Config{})
				require.NoError(err)
				err = gormDB.Where(where.Condition, where.Args...).Find(&found).Error
				require.NoError(err)
				datesWithinRange(t, tc.want, found)
				assert.Equal(tc.want, found)
			}
			{
				// test stdlib
				var found []*user
				tc.opts = append(tc.opts, mql.WithPgPlaceholders())
				where, err = mql.Parse(tc.query, user{}, tc.opts...)
				require.NoError(err)
				d, err := db.SqlDB(testCtx)
				require.NoError(err)
				q := fmt.Sprintf("select * from users where %s", where.Condition)
				rows, err := d.Query(q, where.Args...)
				require.NoError(err)
				defer rows.Close()

				for rows.Next() {
					var u user
					err := rows.Scan(
						&u.ID,
						&u.Name,
						&u.Email,
						&u.Age,
						&u.Birthday,
						&u.CreatedAt,
					)
					require.NoError(err)
					found = append(found, &u)
				}
				require.NoError(rows.Err())
				datesWithinRange(t, tc.want, found)
				assert.Equal(tc.want, found)
			}
		})
	}
}

func datesWithinRange(t *testing.T, want []*user, found []*user) {
	t.Helper()
	assert, require := assert.New(t), require.New(t)
	require.Len(found, len(want), "expected %d and got %d", len(want), len(found))
	for i, u := range found {
		assert.WithinRange(u.CreatedAt, want[i].CreatedAt.Add(-30*time.Second), want[i].CreatedAt.Add(30*time.Second))
		u.CreatedAt = want[i].CreatedAt
	}
}

func pointer[T any](input T) *T {
	ret := input
	return &ret
}
