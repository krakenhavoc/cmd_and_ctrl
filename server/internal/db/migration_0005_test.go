package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

// Tests for migration 0005 (the deck library, ADR 0051 decision 7):
// the new `decks` table, and the second rebuild of `seats` that turns
// deck_id into a real foreign key into it. The interesting case is a
// database an older binary left at v3 with rows in it — seats.deck_id
// has never had a value written into it (0002's and 0003's own
// comments say so), so the rebuild's foreign_key_check should have
// nothing to trip over, and every row and index must survive.

func TestMigration0005FromEmpty(t *testing.T) {
	d := openAtVersion(t, 0)

	for _, table := range []string{"users", "identities", "games", "seats", "invites", "decks"} {
		if n := countRows(t, d, table); n != 0 {
			t.Errorf("%s: %d rows in a fresh database", table, n)
		}
	}
	assertDeckForeignKeys(t, d)
}

func TestMigration0005PreservesAPopulatedV3Database(t *testing.T) {
	d := openAtVersion(t, 3)

	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'Alice', 1, 1)`)
	mustExec(t, d, `INSERT INTO games (id, name, created_by, state, created_at, started_at, ended_at, archived_at, winner_seat)
		VALUES ('g1', 'FNM', 'u1', 'lobby', 1000, NULL, NULL, NULL, NULL),
		       ('g2', 'Old', NULL, 'ended', 2000, 2100, 2200, 2300, 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, user_id, guest_name, bot_tier, deck_id, deck_name, pending_discord_id)
		VALUES ('g1', 0, 'p1', 'u1', 'Alice', NULL, NULL, 'Atraxa', NULL),
		       ('g1', 1, 'p2', NULL, 'Bot', 'heuristic', NULL, NULL, NULL),
		       ('g2', 0, 'p3', NULL, 'Carol', NULL, NULL, 'Krenko', '123456789')`)
	mustExec(t, d, `INSERT INTO invites (token_hash, game_id, kind, created_by, created_at, expires_at, revoked_at)
		VALUES (x'01', 'g1', 'player', 'u1', 1000, NULL, NULL),
		       (x'02', 'g1', 'spectator', NULL, 1000, NULL, NULL),
		       (x'03', 'g2', 'player', NULL, 2000, 9000, 2400)`)

	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v3 -> latest: %v", err)
	}

	for table, want := range map[string]int{"users": 1, "games": 2, "seats": 3, "invites": 3, "decks": 0} {
		if got := countRows(t, d, table); got != want {
			t.Errorf("%s: %d rows, want %d", table, got, want)
		}
	}

	var (
		name     string
		deckName sql.NullString
		userID   sql.NullString
	)
	if err := d.QueryRow(`SELECT guest_name, deck_name, user_id FROM seats WHERE game_id = 'g1' AND seat = 0`).
		Scan(&name, &deckName, &userID); err != nil {
		t.Fatalf("read seat: %v", err)
	}
	if name != "Alice" || deckName.String != "Atraxa" || userID.String != "u1" {
		t.Errorf("seat g1/0 = %q %q %q", name, deckName.String, userID.String)
	}

	assertDeckForeignKeys(t, d)

	// The cascade from games is still wired after the second rebuild.
	mustExec(t, d, `DELETE FROM games WHERE id = 'g1'`)
	for _, table := range []string{"seats", "invites"} {
		var n int
		if err := d.QueryRow(`SELECT COUNT(*) FROM ` + table + ` WHERE game_id = 'g1'`).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if n != 0 {
			t.Errorf("after deleting g1, %s still has %d of its rows (ON DELETE CASCADE lost)", table, n)
		}
	}
}

// TestMigration0005WithDanglingDeckIDRollsBack pins the runner's
// foreign_key_check for the case that matters here: a seat left
// pointing at a deck id that names nothing must still stop the
// migration, even with a v2/v3 schema that never enforced it.
func TestMigration0005WithDanglingDeckIDRollsBack(t *testing.T) {
	d := openAtVersion(t, 3)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'FNM', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, deck_id) VALUES ('g1', 0, 'p1', 'no-such-deck')`)

	err := migrate(context.Background(), d)
	if err == nil || !strings.Contains(err.Error(), "dangling foreign key") {
		t.Fatalf("migrate: got %v, want a dangling-foreign-key refusal", err)
	}
	v, err := currentSchemaVersion(context.Background(), d)
	if err != nil {
		t.Fatalf("currentSchemaVersion: %v", err)
	}
	// migration 0004 (the table host, #1032) sits between this
	// database's v3 starting point and 0005 and has no dangling
	// reference of its own, so migrate() applies it cleanly before
	// reaching — and rolling back — 0005. The version this refusal
	// should leave behind is therefore 4, the last migration that
	// actually committed, not 3.
	if v != 4 {
		t.Errorf("schema version after a refused migration = %d, want 4", v)
	}
	assertForeignKeysOnEveryConn(t, d)
}

// assertDeckForeignKeys checks the shape 0005 promises: decks.owner_id
// references users, seats.deck_id references decks (freshly enforced),
// the other seats/invites/games references from 0003 are unmoved, the
// indexes are back, and the constraint is actually enforced on the
// pool's connections.
func assertDeckForeignKeys(t *testing.T, d *sql.DB) {
	t.Helper()

	if refs := referencesOf(t, d, "decks"); refs["owner_id"] != "users" {
		t.Errorf("decks FKs = %v, want owner_id -> users", refs)
	}
	refs := referencesOf(t, d, "seats")
	if refs["deck_id"] != "decks" {
		t.Errorf("seats FKs = %v, want deck_id -> decks", refs)
	}
	if refs["user_id"] != "users" || refs["game_id"] != "games" {
		t.Errorf("seats FKs = %v, want user_id -> users, game_id -> games", refs)
	}
	if refs := referencesOf(t, d, "invites"); refs["created_by"] != "users" || refs["game_id"] != "games" {
		t.Errorf("invites FKs = %v, want created_by -> users, game_id -> games", refs)
	}
	if refs := referencesOf(t, d, "games"); refs["created_by"] != "users" {
		t.Errorf("games FKs = %v, want created_by -> users", refs)
	}

	if got := strings.Join(indexesOn(t, d, "decks"), ","); got != "decks_owner_id" {
		t.Errorf("decks indexes = %s", got)
	}
	if got := strings.Join(indexesOn(t, d, "seats"), ","); got != "seats_pending_discord_id,seats_user_id" {
		t.Errorf("seats indexes = %s", got)
	}
	var leftovers int
	if err := d.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name LIKE '%\_new' ESCAPE '\'`).Scan(&leftovers); err != nil {
		t.Fatalf("leftovers: %v", err)
	}
	if leftovers != 0 {
		t.Errorf("%d *_new objects left in the schema", leftovers)
	}

	if _, err := d.Exec(`INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
		VALUES ('d1', 'nobody', 'x', 'text', 'y', '[]', 1, 1, 1)`); err == nil {
		t.Error("decks.owner_id accepted a user that does not exist")
	}
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u9', 'Bob', 1, 1)`)
	mustExec(t, d, `INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
		VALUES ('d9', 'u9', 'x', 'text', 'y', '[]', 1, 1, 1)`)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('gz', 'z', 'lobby', 1)`)
	if _, err := d.Exec(`INSERT INTO seats (game_id, seat, player_id, deck_id) VALUES ('gz', 0, 'p', 'no-such-deck')`); err == nil {
		t.Error("seats.deck_id accepted a deck that does not exist")
	}
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, deck_id) VALUES ('gz', 0, 'p', 'd9')`)

	assertForeignKeysOnEveryConn(t, d)
}
