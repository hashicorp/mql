// Copyright IBM Corp. 2023, 2025
// SPDX-License-Identifier: MPL-2.0

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-dbw"
	"github.com/stretchr/testify/require"
)

type user struct {
	ID        uint
	Name      string
	Email     *string
	Age       uint8
	Birthday  *time.Time
	CreatedAt time.Time
}

func testInsertUser(t *testing.T, rw dbw.Writer, u *user) {
	t.Helper()
	testCtx := context.Background()
	require.NoError(t, rw.Create(testCtx, u))
}

const (
	testDbDsn                = "postgresql://go_db:go_db@localhost:9920/go_db?sslmode=disable"
	testCreateTablesPostgres = `
	CREATE TABLE "users" (
		"id" bigserial,
		"name" text,
		"email" text,
		"age" smallint,
		"birthday" timestamptz,
		"created_at" timestamptz,
		PRIMARY KEY ("id")
		)`
)

func testCreateSchema(ctx context.Context, _, url string) error {
	conn, err := dbw.Open(dbw.Postgres, url)
	if err != nil {
		return err
	}
	rw := dbw.New(conn)
	_, err = rw.Exec(context.Background(), testCreateTablesPostgres, nil)
	if err != nil {
		return err
	}
	return nil
}

func setupDB(t *testing.T) *dbw.DB {
	db, _ := dbw.TestSetup(t, dbw.WithTestMigration(testCreateSchema), dbw.WithTestDatabaseUrl(testDbDsn), dbw.WithTestDialect(dbw.Postgres.String()))
	if os.Getenv("DEBUG") != "" {
		db.Debug(true)
	}
	return db
}
