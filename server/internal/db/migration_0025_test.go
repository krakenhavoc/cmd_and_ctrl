package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0025 (a person's own model key for their AI seats,
// ADR 0182): the table's checks, a key row needs a real user, the two
// seat columns arrive empty on old rows, and the hand rollback in
// docs/environments.md.

const modelKeysVersion = 25

// aSealedKey is shaped like a sealed value (version, nonce, ciphertext,
// tag): long enough for the length check, and opaque.
var aSealedKey = make([]byte, 40)

func insertModelKey(d *sql.DB, user string, gamesPerDay, callsPerGame int) error {
	_, err := d.Exec(`INSERT INTO model_keys (user_id, provider, base_url, model, sealed_key, last4, games_per_day, calls_per_game,
		created_at, updated_at) VALUES (?, 'anthropic', '', 'claude-x', ?, 'abcd', ?, ?, 1, 1)`,
		user, aSealedKey, gamesPerDay, callsPerGame)
	return err
}

func TestMigration0025PreservesAPopulatedV24DatabaseWithEmptySeatColumns(t *testing.T) {
	d := openAtVersion(t, modelKeysVersion-1)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'Friday', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, guest_name, bot_tier) VALUES ('g1', 0, 'p1', 'Bot 1', 'heuristic')`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v24 -> latest: %v", err)
	}
	var owner, model sql.NullString
	var tier string
	if err := d.QueryRow(`SELECT bot_tier, bot_key_user_id, bot_key_model FROM seats WHERE game_id = 'g1'`).
		Scan(&tier, &owner, &model); err != nil {
		t.Fatal(err)
	}
	if tier != "heuristic" || owner.Valid || model.Valid {
		t.Errorf("an existing seat: tier %q owner %v model %v; want its tier and both new columns NULL", tier, owner, model)
	}
	if err := insertModelKey(d, "u1", 10, 500); err != nil {
		t.Fatalf("a valid key row: %v", err)
	}
	var dayGames int
	var day, lastErr sql.NullString
	if err := d.QueryRow(`SELECT day, day_games, last_error FROM model_keys WHERE user_id = 'u1'`).Scan(&day, &dayGames, &lastErr); err != nil {
		t.Fatal(err)
	}
	if day.Valid || dayGames != 0 || lastErr.Valid {
		t.Errorf("a new key row: day %v games %d error %v; want no day, 0 and no error", day, dayGames, lastErr)
	}
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0025RefusesABadKeyRow(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	for name, limits := range map[string][2]int{
		"no games": {0, 10}, "too many games": {101, 10}, "no calls": {5, 0}, "too many calls": {5, 5001},
	} {
		if err := insertModelKey(d, "u1", limits[0], limits[1]); err == nil {
			t.Errorf("%s: accepted", name)
			mustExec(t, d, `DELETE FROM model_keys`)
		}
	}
	if err := insertModelKey(d, "nobody", 10, 500); err == nil {
		t.Error("a key row for a user that does not exist was accepted")
	}
	if _, err := d.Exec(`INSERT INTO model_keys (user_id, provider, model, sealed_key, last4, games_per_day, calls_per_game, created_at, updated_at)
		VALUES ('u1', 'anthropic', 'm', x'0102', 'abcd', 1, 1, 1, 1)`); err == nil {
		t.Error("a sealed key too short to be sealed was accepted")
	}
	if _, err := d.Exec(`INSERT INTO model_keys (user_id, provider, model, sealed_key, last4, games_per_day, calls_per_game, created_at, updated_at)
		VALUES ('u1', '', 'm', ?, 'abcd', 1, 1, 1, 1)`, aSealedKey); err == nil {
		t.Error("an empty provider was accepted")
	}
	if err := insertModelKey(d, "u1", 10, 500); err != nil {
		t.Fatal(err)
	}
	if err := insertModelKey(d, "u1", 10, 500); err == nil {
		t.Error("a second key row for one person was accepted")
	}
}

func TestMigration0025RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, modelKeysVersion)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	if err := insertModelKey(d, "u1", 10, 500); err != nil {
		t.Fatal(err)
	}
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'Friday', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, guest_name, bot_tier, bot_key_user_id, bot_key_model)
		VALUES ('g1', 0, 'p1', 'Bot 1', 'assisted', 'u1', 'claude-x')`)
	for _, q := range []string{
		`DROP TABLE model_keys`,
		`ALTER TABLE seats DROP COLUMN bot_key_user_id`,
		`ALTER TABLE seats DROP COLUMN bot_key_model`,
		`DELETE FROM schema_migrations WHERE version = 25`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != modelKeysVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, modelKeysVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	if n := countRows(t, d, "model_keys"); n != 0 {
		t.Errorf("after the round trip model_keys has %d rows, want none", n)
	}
	var owner sql.NullString
	if err := d.QueryRow(`SELECT bot_key_user_id FROM seats WHERE game_id = 'g1'`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if owner.Valid {
		t.Errorf("after the round trip the seat still names a key owner: %v", owner)
	}
}
