package db

import (
	"context"
	"database/sql"
	"testing"
)

// TestMigration0004AddsNullHostColumns: a database an older binary left
// at v3 with a game in it gains host_player_id / host_discord_id, NULL
// on the existing row, and the row is otherwise untouched.
func TestMigration0004AddsNullHostColumns(t *testing.T) {
	d := openAtVersion(t, 3)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'FNM', 'lobby', 1)`)

	if err := migrateTo(context.Background(), d, 4); err != nil {
		t.Fatalf("migrateTo(4): %v", err)
	}
	var (
		name                    string
		hostPlayer, hostDiscord sql.NullString
	)
	if err := d.QueryRow(`SELECT name, host_player_id, host_discord_id FROM games WHERE id = 'g1'`).
		Scan(&name, &hostPlayer, &hostDiscord); err != nil {
		t.Fatalf("select: %v", err)
	}
	if name != "FNM" || hostPlayer.Valid || hostDiscord.Valid {
		t.Errorf("row = %q / %v / %v, want FNM with NULL hosts", name, hostPlayer, hostDiscord)
	}
	mustExec(t, d, `UPDATE games SET host_player_id = 'p', host_discord_id = '1' WHERE id = 'g1'`)
}
