package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0031 (closing idle tables, ADR 0211 §1 and §4):
// games.last_active_at and games.closed_reason are nullable on every
// row old and new, closed_reason takes only the three idle reasons, and
// the hand rollback in docs/environments.md.

// idleTablesVersion is this migration's number.
const idleTablesVersion = 31

func TestMigration0031LeavesExistingGamesOpenAndUntimed(t *testing.T) {
	d := openAtVersion(t, idleTablesVersion-1)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at, results) VALUES ('g1', 'Friday', 'active', 1, 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate to latest: %v", err)
	}
	last, reason := idleColumnsOf(t, d, "g1")
	if last.Valid || reason.Valid {
		t.Errorf("an existing game after the migration: last_active_at %v, closed_reason %v; want both NULL", last, reason)
	}
	var name string
	var results int
	if err := d.QueryRow(`SELECT name, results FROM games WHERE id = 'g1'`).Scan(&name, &results); err != nil {
		t.Fatal(err)
	}
	if name != "Friday" || results != 1 {
		t.Errorf("g1 lost its columns: %q %d", name, results)
	}
}

func TestMigration0031AcceptsOnlyTheIdleReasons(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'A', 'lobby', 1)`)
	for _, ok := range []string{"idle_lobby", "idle_game", "idle_bots"} {
		mustExec(t, d, `UPDATE games SET closed_reason = '`+ok+`', last_active_at = 5 WHERE id = 'g1'`)
	}
	mustExec(t, d, `UPDATE games SET closed_reason = NULL WHERE id = 'g1'`)
	for _, bad := range []string{"admin", "", "IDLE_GAME"} {
		if _, err := d.Exec(`UPDATE games SET closed_reason = ? WHERE id = 'g1'`, bad); err == nil {
			t.Errorf("closed_reason %q accepted", bad)
		}
	}
}

func TestMigration0031RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, idleTablesVersion)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at, last_active_at, closed_reason)
	                VALUES ('g1', 'A', 'active', 1, 9, 'idle_game')`)
	for _, q := range []string{
		`ALTER TABLE games DROP COLUMN closed_reason`,
		`ALTER TABLE games DROP COLUMN last_active_at`,
		`DELETE FROM schema_migrations WHERE version = 31`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != idleTablesVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, idleTablesVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	last, reason := idleColumnsOf(t, d, "g1")
	if last.Valid || reason.Valid {
		t.Errorf("after the round trip: last_active_at %v, closed_reason %v; want both NULL", last, reason)
	}
}

func idleColumnsOf(t *testing.T, d *sql.DB, id string) (sql.NullInt64, sql.NullString) {
	t.Helper()
	var last sql.NullInt64
	var reason sql.NullString
	if err := d.QueryRow(`SELECT last_active_at, closed_reason FROM games WHERE id = ?`, id).Scan(&last, &reason); err != nil {
		t.Fatalf("read the idle columns of %s: %v", id, err)
	}
	return last, reason
}
