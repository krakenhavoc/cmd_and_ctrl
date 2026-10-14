package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0026 (a bot standing in for an away player, ADR
// 0177 §4): the two seat columns arrive empty on old rows, and the hand
// rollback in docs/environments.md.

const standInVersion = 26

func TestMigration0026PreservesAPopulatedV25DatabaseWithEmptyStandIn(t *testing.T) {
	d := openAtVersion(t, standInVersion-1)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'Friday', 'active', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, guest_name) VALUES ('g1', 0, 'p1', 'Ana')`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v25 -> latest: %v", err)
	}
	var tier sql.NullString
	var since sql.NullInt64
	if err := d.QueryRow(`SELECT stand_in, stand_in_since FROM seats WHERE game_id = 'g1'`).Scan(&tier, &since); err != nil {
		t.Fatal(err)
	}
	if tier.Valid || since.Valid {
		t.Errorf("an existing seat: stand_in %v since %v; want both NULL", tier, since)
	}
	mustExec(t, d, `UPDATE seats SET stand_in = 'heuristic', stand_in_since = 5 WHERE game_id = 'g1'`)
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0026RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, standInVersion)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'Friday', 'active', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, guest_name, stand_in, stand_in_since)
		VALUES ('g1', 0, 'p1', 'Ana', 'heuristic', 5)`)
	for _, q := range []string{
		`ALTER TABLE seats DROP COLUMN stand_in`,
		`ALTER TABLE seats DROP COLUMN stand_in_since`,
		`DELETE FROM schema_migrations WHERE version = 26`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != standInVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, standInVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	var tier sql.NullString
	if err := d.QueryRow(`SELECT stand_in FROM seats WHERE game_id = 'g1'`).Scan(&tier); err != nil {
		t.Fatal(err)
	}
	if tier.Valid {
		t.Errorf("after the round trip the seat still has a stand-in: %v", tier)
	}
}
