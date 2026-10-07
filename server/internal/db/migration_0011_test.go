package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0011 (users.playmat_id, ADR 0128): one nullable
// column, NULL on every existing row, no backfill.

func TestMigration0011FromEmpty(t *testing.T) {
	d := openAtVersion(t, 0)
	assertPlaymatShape(t, d)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, playmat_id) VALUES ('u2', 'B', 1, 1, 'mat-2')`)
	if got := playmatID(t, d, "u1"); got != nil {
		t.Errorf("a new user has playmat_id %q, want NULL", *got)
	}
	if got := playmatID(t, d, "u2"); got == nil || *got != "mat-2" {
		t.Errorf("u2 playmat_id = %v, want mat-2", got)
	}
}

// Every user that exists when this deploys reads NULL, and nothing
// else about the row moves.
func TestMigration0011PreservesAPopulatedV10Database(t *testing.T) {
	d := openAtVersion(t, 10)
	mustExec(t, d, `INSERT INTO users (id, display_name, avatar_url, created_at, last_seen_at)
		VALUES ('u1', 'Alice', '/avatars/1/a.png', 5, 6), ('u2', 'Bob', NULL, 7, 8)`)

	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v10 -> latest: %v", err)
	}
	if n := countRows(t, d, "users"); n != 2 {
		t.Errorf("users: %d rows, want 2", n)
	}
	for _, u := range []string{"u1", "u2"} {
		if got := playmatID(t, d, u); got != nil {
			t.Errorf("%s playmat_id = %q after the migration, want NULL", u, *got)
		}
	}
	var name string
	var avatar sql.NullString
	var seen int64
	if err := d.QueryRow(`SELECT display_name, avatar_url, last_seen_at FROM users WHERE id = 'u1'`).Scan(&name, &avatar, &seen); err != nil {
		t.Fatal(err)
	}
	if name != "Alice" || avatar.String != "/avatars/1/a.png" || seen != 6 {
		t.Errorf("u1 lost its columns: %q %v %d", name, avatar, seen)
	}
	assertPlaymatShape(t, d)
}

// The hand rollback documented in the migration leaves a database that
// rolls forward cleanly.
func TestMigration0011RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 11)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, playmat_id) VALUES ('u1', 'A', 1, 1, 'mat-1')`)
	for _, q := range []string{
		`ALTER TABLE users DROP COLUMN playmat_id`,
		`DELETE FROM schema_migrations WHERE version = 11`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 10 {
		t.Fatalf("after the rollback the version is %d, want 10", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	assertPlaymatShape(t, d)
	if got := playmatID(t, d, "u1"); got != nil {
		t.Errorf("after the round trip playmat_id = %q, want NULL (no backfill)", *got)
	}
}

func playmatID(t *testing.T, d *sql.DB, user string) *string {
	t.Helper()
	var c *string
	if err := d.QueryRow(`SELECT playmat_id FROM users WHERE id = ?`, user).Scan(&c); err != nil {
		t.Fatalf("read playmat_id for %s: %v", user, err)
	}
	return c
}

func assertPlaymatShape(t *testing.T, d *sql.DB) {
	t.Helper()
	var (
		typ     string
		notNull int
		dflt    *string
	)
	err := d.QueryRow(`SELECT type, "notnull", dflt_value FROM pragma_table_info('users') WHERE name = 'playmat_id'`).Scan(&typ, &notNull, &dflt)
	if err != nil {
		t.Fatalf("users.playmat_id missing: %v", err)
	}
	if typ != "TEXT" || notNull != 0 || dflt != nil {
		t.Errorf("users.playmat_id = %s notnull=%d default=%v, want a nullable TEXT with no default", typ, notNull, dflt)
	}
	assertForeignKeysOnEveryConn(t, d)
}
