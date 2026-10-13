package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0019 (the sign-up gate's last check, ADR 0156 §1
// item 7): two nullable users columns, NULL ("unchecked") on every row
// old and new, a result outside the three answers refused, and the hand
// rollback in docs/environments.md.

func TestMigration0019PreservesAPopulatedV18DatabaseUnchecked(t *testing.T) {
	d := openAtVersion(t, 18)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, sessions_invalid_before, admin_mode_at)
		VALUES ('u1', 'Owner', 1, 2, 3, 4)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v18 -> latest: %v", err)
	}
	assertAccessCheckShape(t, d)
	at, result := accessCheck(t, d, "u1")
	if at.Valid || result.Valid {
		t.Errorf("an existing user after the migration: checked_at %v, result %v; want both NULL", at, result)
	}
	var watermark, mode int64
	if err := d.QueryRow(`SELECT sessions_invalid_before, admin_mode_at FROM users WHERE id = 'u1'`).Scan(&watermark, &mode); err != nil {
		t.Fatal(err)
	}
	if watermark != 3 || mode != 4 {
		t.Errorf("u1 lost its columns: watermark %d, admin_mode_at %d", watermark, mode)
	}
}

func TestMigration0019AcceptsOnlyTheThreeAnswers(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	for _, ok := range []string{"admitted", "not_member", "no_role"} {
		mustExec(t, d, `UPDATE users SET access_result = ?, access_checked_at = 5 WHERE id = 'u1'`, ok)
	}
	mustExec(t, d, `UPDATE users SET access_result = NULL, access_checked_at = NULL WHERE id = 'u1'`)
	for _, bad := range []string{"refused", "outage", ""} {
		if _, err := d.Exec(`UPDATE users SET access_result = ? WHERE id = 'u1'`, bad); err == nil {
			t.Errorf("access_result %q accepted", bad)
		}
	}
}

func TestMigration0019RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 19)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, access_checked_at, access_result)
		VALUES ('u1', 'A', 1, 1, 1760000000000, 'admitted')`)
	for _, q := range []string{
		`ALTER TABLE users DROP COLUMN access_checked_at`,
		`ALTER TABLE users DROP COLUMN access_result`,
		`DELETE FROM schema_migrations WHERE version = 19`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 18 {
		t.Fatalf("after the rollback the version is %d, want 18", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	assertAccessCheckShape(t, d)
	if at, result := accessCheck(t, d, "u1"); at.Valid || result.Valid {
		t.Errorf("after the round trip: %v %v, want unchecked", at, result)
	}
}

func accessCheck(t *testing.T, d *sql.DB, id string) (sql.NullInt64, sql.NullString) {
	t.Helper()
	var (
		at     sql.NullInt64
		result sql.NullString
	)
	if err := d.QueryRow(`SELECT access_checked_at, access_result FROM users WHERE id = ?`, id).Scan(&at, &result); err != nil {
		t.Fatalf("read access check for %s: %v", id, err)
	}
	return at, result
}

func assertAccessCheckShape(t *testing.T, d *sql.DB) {
	t.Helper()
	for col, want := range map[string]string{"access_checked_at": "INTEGER", "access_result": "TEXT"} {
		var (
			typ     string
			notNull int
		)
		if err := d.QueryRow(`SELECT type, "notnull" FROM pragma_table_info('users') WHERE name = ?`, col).Scan(&typ, &notNull); err != nil {
			t.Fatalf("users.%s missing: %v", col, err)
		}
		if typ != want || notNull != 0 {
			t.Errorf("users.%s = %s notnull=%d, want nullable %s", col, typ, notNull, want)
		}
	}
	assertForeignKeysOnEveryConn(t, d)
}
