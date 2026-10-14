package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0024 (a table's opt-in to the Discord results
// feed, ADR 0174 §1): games.results defaults to 0 on every row old and
// new and takes only 0 or 1, and the hand rollback in
// docs/environments.md.

func TestMigration0024PreservesAPopulatedV23DatabaseWithResultsOff(t *testing.T) {
	d := openAtVersion(t, 23)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at, lfg) VALUES ('g1', 'Friday', 'ended', 1, 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v23 -> latest: %v", err)
	}
	assertResultsShape(t, d)
	if got := resultsOf(t, d, "g1"); got != 0 {
		t.Errorf("an existing game after the migration: results %d, want 0", got)
	}
	var name string
	var lfg int
	if err := d.QueryRow(`SELECT name, lfg FROM games WHERE id = 'g1'`).Scan(&name, &lfg); err != nil {
		t.Fatal(err)
	}
	if name != "Friday" || lfg != 1 {
		t.Errorf("g1 lost its columns: %q %d", name, lfg)
	}
}

func TestMigration0024AcceptsOnlyOnOrOff(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'A', 'lobby', 1)`)
	mustExec(t, d, `UPDATE games SET results = 1 WHERE id = 'g1'`)
	mustExec(t, d, `UPDATE games SET results = 0 WHERE id = 'g1'`)
	for _, bad := range []any{2, -1, "yes"} {
		if _, err := d.Exec(`UPDATE games SET results = ? WHERE id = 'g1'`, bad); err == nil {
			t.Errorf("results %v accepted", bad)
		}
	}
	if _, err := d.Exec(`UPDATE games SET results = NULL WHERE id = 'g1'`); err == nil {
		t.Error("results NULL accepted")
	}
}

func TestMigration0024RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 24)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at, results) VALUES ('g1', 'A', 'lobby', 1, 1)`)
	for _, q := range []string{
		`ALTER TABLE games DROP COLUMN results`,
		`DELETE FROM schema_migrations WHERE version = 24`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 23 {
		t.Fatalf("after the rollback the version is %d, want 23", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	assertResultsShape(t, d)
	if got := resultsOf(t, d, "g1"); got != 0 {
		t.Errorf("after the round trip results is %d, want off", got)
	}
}

func resultsOf(t *testing.T, d *sql.DB, id string) int {
	t.Helper()
	var n int
	if err := d.QueryRow(`SELECT results FROM games WHERE id = ?`, id).Scan(&n); err != nil {
		t.Fatalf("read games.results for %s: %v", id, err)
	}
	return n
}

func assertResultsShape(t *testing.T, d *sql.DB) {
	t.Helper()
	var (
		typ     string
		notNull int
	)
	if err := d.QueryRow(`SELECT type, "notnull" FROM pragma_table_info('games') WHERE name = 'results'`).Scan(&typ, &notNull); err != nil {
		t.Fatalf("games.results missing: %v", err)
	}
	if typ != "INTEGER" || notNull != 1 {
		t.Errorf("games.results = %s notnull=%d, want INTEGER notnull=1", typ, notNull)
	}
	assertForeignKeysOnEveryConn(t, d)
}
