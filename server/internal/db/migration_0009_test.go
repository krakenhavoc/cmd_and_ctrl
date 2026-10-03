package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0009 (admin mode, ADR 0112 §2):
// users.admin_mode_at, 0 (player mode) on every row old and new.

func TestMigration0009FromEmpty(t *testing.T) {
	d := openAtVersion(t, 0)
	assertAdminModeShape(t, d)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	if got := adminModeAt(t, d, "u1"); got != 0 {
		t.Errorf("a new user's admin_mode_at = %d, want 0 (player mode)", got)
	}
}

// Every person who exists when this deploys starts in player mode, the
// owner included (ADR 0112 §2 item 1).
func TestMigration0009PreservesAPopulatedV8DatabaseInPlayerMode(t *testing.T) {
	d := openAtVersion(t, 8)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, sessions_invalid_before, last_deck)
		VALUES ('u1', 'Owner', 1, 2, 3, '{"kind":"prebuilt","id":"x"}'),
		       ('u2', 'Player', 4, 5, 0, NULL)`)

	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v8 -> latest: %v", err)
	}
	if n := countRows(t, d, "users"); n != 2 {
		t.Errorf("users: %d rows, want 2", n)
	}
	for _, id := range []string{"u1", "u2"} {
		if got := adminModeAt(t, d, id); got != 0 {
			t.Errorf("%s admin_mode_at = %d after the migration, want 0", id, got)
		}
	}
	var watermark int64
	var lastDeck *string
	if err := d.QueryRow(`SELECT sessions_invalid_before, last_deck FROM users WHERE id = 'u1'`).Scan(&watermark, &lastDeck); err != nil {
		t.Fatal(err)
	}
	if watermark != 3 || lastDeck == nil {
		t.Errorf("u1 lost its columns: watermark %d, last_deck %v", watermark, lastDeck)
	}
	assertAdminModeShape(t, d)
}

// The column refuses NULL: a reader can trust 0 means off.
func TestMigration0009AdminModeAtIsNotNull(t *testing.T) {
	d := openAtVersion(t, 0)
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at, admin_mode_at) VALUES ('u1', 'A', 1, 1, NULL)`); err == nil {
		t.Error("users.admin_mode_at accepted NULL")
	}
}

// An older binary refuses a v9 database, and the hand rollback
// documented in docs/environments.md leaves a database that rolls
// forward cleanly, in player mode.
func TestMigration0009RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 9)
	if v, err := currentSchemaVersion(ctx, d); err != nil || v < 9 {
		t.Fatalf("version = %d (%v), want >= 9", v, err)
	}
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, admin_mode_at) VALUES ('u1', 'A', 1, 1, 1759400000000)`)
	for _, q := range []string{
		`ALTER TABLE users DROP COLUMN admin_mode_at`,
		`DELETE FROM schema_migrations WHERE version = 9`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 8 {
		t.Fatalf("after the rollback the version is %d, want 8", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	assertAdminModeShape(t, d)
	if got := adminModeAt(t, d, "u1"); got != 0 {
		t.Errorf("after the round trip admin_mode_at = %d, want 0: the person is back in player mode", got)
	}
}

func adminModeAt(t *testing.T, d *sql.DB, id string) int64 {
	t.Helper()
	var at int64
	if err := d.QueryRow(`SELECT admin_mode_at FROM users WHERE id = ?`, id).Scan(&at); err != nil {
		t.Fatalf("read admin_mode_at for %s: %v", id, err)
	}
	return at
}

func assertAdminModeShape(t *testing.T, d *sql.DB) {
	t.Helper()
	var (
		typ     string
		notNull int
		dflt    *string
	)
	err := d.QueryRow(`SELECT type, "notnull", dflt_value FROM pragma_table_info('users') WHERE name = 'admin_mode_at'`).Scan(&typ, &notNull, &dflt)
	if err != nil {
		t.Fatalf("users.admin_mode_at missing: %v", err)
	}
	if typ != "INTEGER" || notNull != 1 || dflt == nil || *dflt != "0" {
		t.Errorf("users.admin_mode_at = %s notnull=%d default=%v, want INTEGER NOT NULL DEFAULT 0", typ, notNull, dflt)
	}
	assertForeignKeysOnEveryConn(t, d)
}
