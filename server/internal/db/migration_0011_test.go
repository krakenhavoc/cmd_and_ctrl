package db

// Tests for migration 0011 (ADR 0128): user_playmats, its checks, and
// the by-hand rollback docs/environments.md gives.

import (
	"context"
	"testing"
)

func TestMigration0011CreatesUserPlaymats(t *testing.T) {
	d := openAtVersion(t, 11)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO user_playmats (user_id, file, wash, updated_at) VALUES ('u1', 'abc.png', 60, 1)`)
	var wash int
	if err := d.QueryRow(`SELECT wash FROM user_playmats WHERE user_id = 'u1'`).Scan(&wash); err != nil || wash != 60 {
		t.Fatalf("read back wash = %d, %v", wash, err)
	}
	// The checks are the backstop for the store's own.
	for _, q := range []string{
		`UPDATE user_playmats SET wash = 29 WHERE user_id = 'u1'`,
		`UPDATE user_playmats SET wash = 91 WHERE user_id = 'u1'`,
		`UPDATE user_playmats SET file = '' WHERE user_id = 'u1'`,
	} {
		if _, err := d.Exec(q); err == nil {
			t.Errorf("%s: want a CHECK failure", q)
		}
	}
}

func TestMigration0011RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 11)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO user_playmats (user_id, file, wash, updated_at) VALUES ('u1', 'abc.png', 60, 1)`)
	for _, q := range []string{
		`DROP TABLE user_playmats`,
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
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM user_playmats`).Scan(&n); err != nil || n != 0 {
		t.Errorf("after the round trip user_playmats has %d rows (%v), want an empty table", n, err)
	}
}
