package db

import (
	"context"
	"testing"
)

// Tests for migration 0016 (table_blocks and player_reports, #2984): a
// v15 database gains both tables, a block is one row per (table, kind,
// subject), a block leaves with its game, and a report does not.

func TestMigration0016AddsBlocksAndReports(t *testing.T) {
	d := openAtVersion(t, 15)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'Table', 'lobby', 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v15 -> latest: %v", err)
	}

	mustExec(t, d, `INSERT INTO table_blocks (game_id, kind, subject, created_at, entry_id) VALUES ('g1', 'name', 'bob', 1, 'e1')`)
	mustExec(t, d, `INSERT INTO table_blocks (game_id, kind, subject, created_at, entry_id) VALUES ('g1', 'user', 'bob', 1, 'e1')`)
	if _, err := d.Exec(`INSERT INTO table_blocks (game_id, kind, subject, created_at, entry_id) VALUES ('g1', 'name', 'bob', 2, 'e2')`); err == nil {
		t.Error("the same block twice was accepted")
	}
	if _, err := d.Exec(`INSERT INTO table_blocks (game_id, kind, subject, created_at, entry_id) VALUES ('g1', 'ip', '1.2.3.4', 1, 'e3')`); err == nil {
		t.Error("a block kind outside user / discord / name / spectator was accepted")
	}
	if _, err := d.Exec(`INSERT INTO table_blocks (game_id, kind, subject, created_at, entry_id) VALUES ('nope', 'name', 'x', 1, 'e4')`); err == nil {
		t.Error("a block on a table that does not exist was accepted")
	}

	mustExec(t, d, `INSERT INTO player_reports (id, created_at, game_id, game_name, reporter_player_id, reporter_name,
		reported_player_id, reported_name, reason, excerpt) VALUES ('r1', 1, 'g1', 'Table', 'p1', 'A', 'p2', 'B', 'rude', '[]')`)

	mustExec(t, d, `DELETE FROM games WHERE id = 'g1'`)
	var blocks, reports int
	if err := d.QueryRow(`SELECT COUNT(*) FROM table_blocks`).Scan(&blocks); err != nil || blocks != 0 {
		t.Errorf("blocks after the table went: %d, %v; want 0 (ON DELETE CASCADE)", blocks, err)
	}
	if err := d.QueryRow(`SELECT COUNT(*) FROM player_reports`).Scan(&reports); err != nil || reports != 1 {
		t.Errorf("reports after the table went: %d, %v; want 1 (a report outlives its table)", reports, err)
	}
}
