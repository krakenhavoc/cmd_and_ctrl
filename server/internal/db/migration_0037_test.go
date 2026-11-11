package db

import (
	"context"
	"testing"
)

// Tests for migration 0037 (the games of a match, ADR 0222 §3): one new
// table that goes with its game's statistics, and the hand rollback in
// docs/environments.md.

// matchGamesVersion is this migration's number.
const matchGamesVersion = 37

const insertStatsForMatch = `INSERT INTO game_stats (game_id, stats_version, source, log_complete, turns, rounds, starting_seat, computed_at)
	VALUES ('g1', 2, 'live', 1, 9, 5, 0, 1)`

func TestMigration0037MatchGamesGoWithTheirStats(t *testing.T) {
	d := openAtVersion(t, matchGamesVersion-1)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'A match', 'ended', 1)`)
	mustExec(t, d, insertStatsForMatch)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate to latest: %v", err)
	}
	mustExec(t, d, `INSERT INTO match_games (game_id, game_no, outcome, winner_seat, cause, starting_seat, turns)
		VALUES ('g1', 1, 'win', 0, 'last_standing', 0, 7), ('g1', 2, 'draw', NULL, 'all_lost', 1, 9)`)
	if _, err := d.Exec(`INSERT INTO match_games (game_id, game_no, outcome, cause, starting_seat, turns)
		VALUES ('g1', 2, 'win', 'last_standing', 0, 3)`); err == nil {
		t.Error("a second row for game 2 of one match was accepted")
	}
	if _, err := d.Exec(`INSERT INTO match_games (game_id, game_no, outcome, cause, starting_seat, turns)
		VALUES ('g1', 3, 'loss', 'last_standing', 0, 3)`); err == nil {
		t.Error("an outcome that is neither a win nor a draw was accepted")
	}
	// Deleting the table deletes its statistics, and the games with them.
	mustExec(t, d, `DELETE FROM games WHERE id = 'g1'`)
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM match_games`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("%d match_games rows outlived their game", n)
	}
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0037RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, matchGamesVersion)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'A match', 'ended', 1)`)
	mustExec(t, d, insertStatsForMatch)
	mustExec(t, d, `INSERT INTO match_games (game_id, game_no, outcome, winner_seat, cause, starting_seat, turns)
		VALUES ('g1', 1, 'win', 0, 'last_standing', 0, 7)`)
	mustExec(t, d, `DROP TABLE match_games`)
	mustExec(t, d, `DELETE FROM schema_migrations WHERE version = ?`, matchGamesVersion)
	if v, _ := currentSchemaVersion(ctx, d); v != matchGamesVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, matchGamesVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("roll forward again: %v", err)
	}
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM match_games`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("rolled forward: %d rows; want the table back, empty", n)
	}
	var turns int
	if err := d.QueryRow(`SELECT turns FROM game_stats WHERE game_id = 'g1'`).Scan(&turns); err != nil || turns != 9 {
		t.Errorf("the game's statistics did not survive the rollback: turns %d, err %v", turns, err)
	}
}
