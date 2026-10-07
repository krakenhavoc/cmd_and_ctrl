package db

import (
	"context"
	"testing"
)

// Tests for migration 0012 (users.playmat_wash, ADR 0128 amendment):
// one nullable column, NULL on every existing row, no backfill.

func TestMigration0012AddsANullableWash(t *testing.T) {
	d := openAtVersion(t, 12)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	var (
		typ     string
		notNull int
		dflt    *string
	)
	if err := d.QueryRow(`SELECT type, "notnull", dflt_value FROM pragma_table_info('users') WHERE name = 'playmat_wash'`).Scan(&typ, &notNull, &dflt); err != nil {
		t.Fatalf("users.playmat_wash missing: %v", err)
	}
	if typ != "INTEGER" || notNull != 0 || dflt != nil {
		t.Errorf("users.playmat_wash = %s notnull=%d default=%v, want a nullable INTEGER with no default", typ, notNull, dflt)
	}
	var w *int
	if err := d.QueryRow(`SELECT playmat_wash FROM users WHERE id = 'u1'`).Scan(&w); err != nil || w != nil {
		t.Errorf("a new user's wash = %v (%v), want NULL", w, err)
	}
}

func TestMigration0012RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 12)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, playmat_wash) VALUES ('u1', 'A', 1, 1, 70)`)
	for _, q := range []string{
		`ALTER TABLE users DROP COLUMN playmat_wash`,
		`DELETE FROM schema_migrations WHERE version = 12`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 11 {
		t.Fatalf("after the rollback the version is %d, want 11", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	var w *int
	if err := d.QueryRow(`SELECT playmat_wash FROM users WHERE id = 'u1'`).Scan(&w); err != nil || w != nil {
		t.Errorf("after the round trip wash = %v (%v), want NULL (no backfill)", w, err)
	}
}
