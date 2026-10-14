package db

import (
	"context"
	"testing"
)

// Tests for migration 0022 (Discord DM notification preferences, ADR
// 0162 §5): the new table's defaults are every switch off, a switch
// takes only 0 or 1, the row needs a real user, and the hand rollback
// in docs/environments.md.

// notifyPrefsVersion is this migration's number.
const notifyPrefsVersion = 22

func TestMigration0022AddsNotificationPrefsOffByDefault(t *testing.T) {
	d := openAtVersion(t, notifyPrefsVersion-1)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate to latest: %v", err)
	}
	mustExec(t, d, `INSERT INTO notification_prefs (user_id, updated_at) VALUES ('u1', 5)`)
	var turn, ready, filled int
	var failed *int64
	if err := d.QueryRow(`SELECT turn, table_ready, table_filled, dm_failed_at FROM notification_prefs WHERE user_id = 'u1'`).
		Scan(&turn, &ready, &filled, &failed); err != nil {
		t.Fatal(err)
	}
	if turn != 0 || ready != 0 || filled != 0 || failed != nil {
		t.Errorf("a new row: turn %d ready %d filled %d failed %v; want every switch off and no failure", turn, ready, filled, failed)
	}
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0022AcceptsOnlyOnOrOffAndARealUser(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO notification_prefs (user_id, turn, table_ready, table_filled, updated_at) VALUES ('u1', 1, 1, 1, 5)`)
	for _, col := range []string{"turn", "table_ready", "table_filled"} {
		for _, bad := range []any{2, -1, "yes", nil} {
			if _, err := d.Exec(`UPDATE notification_prefs SET `+col+` = ? WHERE user_id = 'u1'`, bad); err == nil {
				t.Errorf("%s = %v accepted", col, bad)
			}
		}
	}
	if _, err := d.Exec(`INSERT INTO notification_prefs (user_id, updated_at) VALUES ('nobody', 5)`); err == nil {
		t.Error("a row for a user that does not exist was accepted")
	}
}

func TestMigration0022RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, notifyPrefsVersion)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO notification_prefs (user_id, turn, updated_at) VALUES ('u1', 1, 5)`)
	mustExec(t, d, `DROP TABLE notification_prefs`)
	mustExec(t, d, `DELETE FROM schema_migrations WHERE version = ?`, notifyPrefsVersion)
	if v, _ := currentSchemaVersion(ctx, d); v != notifyPrefsVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, notifyPrefsVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	if n := countRows(t, d, "notification_prefs"); n != 0 {
		t.Errorf("after the round trip notification_prefs has %d rows, want none", n)
	}
}
