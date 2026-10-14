package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0022 (game statistics, ADR 0164 §3): the three
// tables take a record, refuse values outside their CHECKs, go with
// their game, and the hand rollback in docs/environments.md.

func seedStats(t *testing.T, d *sql.DB) {
	t.Helper()
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'Friday', 'ended', 1)`)
	mustExec(t, d, `INSERT INTO game_stats (game_id, stats_version, source, log_complete, turns, rounds, starting_seat, computed_at)
		VALUES ('g1', 1, 'live', 1, 9, 3, 0, 1)`)
	mustExec(t, d, `INSERT INTO seat_stats (game_id, seat, result, place, eliminated_turn, conceded, final_life,
		life_gained, life_lost, damage_to_players, combat_damage_to_players, commander_damage_dealt,
		commander_damage_taken, spells_cast, cards_drawn, attacks, tokens_created, mulligans, commanders)
		VALUES ('g1', 0, 'win', 1, NULL, 0, 31, 6, 15, 40, 30, 21, 0, 12, 14, 9, 3, 1, '["Atraxa, Praetors'' Voice"]')`)
	mustExec(t, d, `INSERT INTO seat_card_stats (game_id, seat, name, scryfall_id, casts, first_cast_turn)
		VALUES ('g1', 0, 'Sol Ring', NULL, 1, 1)`)
}

func TestMigration0022TakesARecordAndItGoesWithItsGame(t *testing.T) {
	d := openAtVersion(t, 0)
	seedStats(t, d)
	for _, table := range []string{"game_stats", "seat_stats", "seat_card_stats"} {
		if n := countRows(t, d, table); n != 1 {
			t.Errorf("%s has %d rows, want 1", table, n)
		}
	}
	mustExec(t, d, `DELETE FROM games WHERE id = 'g1'`)
	for _, table := range []string{"game_stats", "seat_stats", "seat_card_stats"} {
		if n := countRows(t, d, table); n != 0 {
			t.Errorf("%s kept %d rows after its game was deleted", table, n)
		}
	}
}

func TestMigration0022RefusesValuesOutsideItsChecks(t *testing.T) {
	d := openAtVersion(t, 0)
	seedStats(t, d)
	for _, q := range []string{
		`UPDATE game_stats SET source = 'guess' WHERE game_id = 'g1'`,
		`UPDATE game_stats SET log_complete = 2 WHERE game_id = 'g1'`,
		`UPDATE seat_stats SET result = 'victory' WHERE game_id = 'g1'`,
		`UPDATE seat_stats SET conceded = 3 WHERE game_id = 'g1'`,
		// A record for a game with no games row.
		`INSERT INTO game_stats (game_id, stats_version, source, log_complete, turns, rounds, computed_at) VALUES ('nope', 1, 'live', 1, 0, 0, 1)`,
		// A card line for a seat with no seat line.
		`INSERT INTO seat_card_stats (game_id, seat, name, casts, first_cast_turn) VALUES ('g1', 3, 'Sol Ring', 1, 1)`,
	} {
		if _, err := d.Exec(q); err == nil {
			t.Errorf("accepted: %s", q)
		}
	}
}

func TestMigration0022RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 22)
	seedStats(t, d)
	for _, q := range []string{
		`DROP TABLE seat_card_stats`,
		`DROP TABLE seat_stats`,
		`DROP TABLE game_stats`,
		`DELETE FROM schema_migrations WHERE version = 22`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v >= 22 {
		t.Fatalf("after the rollback the version is %d, want below 22", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	if n := countRows(t, d, "game_stats"); n != 0 {
		t.Errorf("rolled forward with %d game_stats rows, want an empty table", n)
	}
	// The game itself survived the round trip.
	if n := countRows(t, d, "games"); n != 1 {
		t.Errorf("games has %d rows, want 1", n)
	}
}
