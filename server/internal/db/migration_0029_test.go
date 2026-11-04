package db

import (
	"context"
	"testing"
)

// Tests for migration 0029 (the advanced-stats setting, ADR 0217 §2):
// a new row follows the default, the switch takes only 0, 1 or NULL
// with a matching setter, the row needs a real user, the server
// default has at most one row, and the hand rollback in
// docs/environments.md.

// statsPrefsVersion is this migration's number.
const statsPrefsVersion = 29

func TestMigration0029AddsStatsPrefsFollowingTheDefault(t *testing.T) {
	d := openAtVersion(t, statsPrefsVersion-1)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate to latest: %v", err)
	}
	mustExec(t, d, `INSERT INTO stats_prefs (user_id, updated_at) VALUES ('u1', 5)`)
	var adv *int
	var by *string
	var pods int
	if err := d.QueryRow(`SELECT advanced, advanced_set_by, share_pods FROM stats_prefs WHERE user_id = 'u1'`).
		Scan(&adv, &by, &pods); err != nil {
		t.Fatal(err)
	}
	if adv != nil || by != nil || pods != 0 {
		t.Errorf("a new row: advanced %v set_by %v share_pods %d; want the default and no sharing", adv, by, pods)
	}
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0029Constraints(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO stats_prefs (user_id, advanced, advanced_set_by, updated_at) VALUES ('u1', 1, 'self', 5)`)
	for _, bad := range []string{
		`UPDATE stats_prefs SET advanced = 2`,
		`UPDATE stats_prefs SET advanced_set_by = 'bot'`,
		`UPDATE stats_prefs SET advanced = NULL`,
		`UPDATE stats_prefs SET advanced_set_by = NULL`,
		`UPDATE stats_prefs SET share_pods = 2`,
		`INSERT INTO stats_prefs (user_id, updated_at) VALUES ('nobody', 5)`,
		`INSERT INTO stats_settings (id, advanced_default, updated_at) VALUES (2, 1, 5)`,
		`INSERT INTO stats_settings (id, advanced_default, updated_at) VALUES (1, 3, 5)`,
	} {
		if _, err := d.Exec(bad); err == nil {
			t.Errorf("accepted: %s", bad)
		}
	}
	mustExec(t, d, `UPDATE stats_prefs SET advanced = NULL, advanced_set_by = NULL`)
	mustExec(t, d, `INSERT INTO stats_settings (id, advanced_default, updated_at) VALUES (1, 1, 5)`)
}

func TestMigration0029RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, statsPrefsVersion)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO stats_prefs (user_id, advanced, advanced_set_by, updated_at) VALUES ('u1', 1, 'admin', 5)`)
	mustExec(t, d, `DROP TABLE stats_prefs`)
	mustExec(t, d, `DROP TABLE stats_settings`)
	mustExec(t, d, `DELETE FROM schema_migrations WHERE version = ?`, statsPrefsVersion)
	if v, _ := currentSchemaVersion(ctx, d); v != statsPrefsVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, statsPrefsVersion-1)
	}
}
