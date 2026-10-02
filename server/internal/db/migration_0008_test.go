package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

// Tests for migration 0008 (remember me, ADR 0110): user_settings,
// table_setups, users.last_deck, decks.source_url, and the third
// rebuild of `seats`, which gives deck_id ON DELETE SET NULL.

func TestMigration0008FromEmpty(t *testing.T) {
	d := openAtVersion(t, 0)
	for _, table := range []string{"user_settings", "table_setups"} {
		if n := countRows(t, d, table); n != 0 {
			t.Errorf("%s: %d rows in a fresh database", table, n)
		}
	}
	assertRememberMeShape(t, d)
}

func TestMigration0008PreservesAPopulatedV7Database(t *testing.T) {
	d := openAtVersion(t, 7)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'Alice', 1, 1)`)
	mustExec(t, d, `INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
		VALUES ('d1', 'u1', 'Atraxa', 'text', '1 Atraxa', '[]', 100, 1, 1)`)
	mustExec(t, d, `INSERT INTO games (id, name, created_by, state, created_at) VALUES ('g1', 'FNM', 'u1', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, user_id, guest_name, bot_tier, deck_id, deck_name, pending_discord_id)
		VALUES ('g1', 0, 'p1', 'u1', 'Alice', NULL, 'd1', 'Atraxa', NULL),
		       ('g1', 1, 'p2', NULL, 'Bot', 'heuristic', NULL, NULL, '123')`)

	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v7 -> latest: %v", err)
	}

	for table, want := range map[string]int{"users": 1, "decks": 1, "games": 1, "seats": 2, "user_settings": 0, "table_setups": 0} {
		if got := countRows(t, d, table); got != want {
			t.Errorf("%s: %d rows, want %d", table, got, want)
		}
	}
	var deckID, deckName, pending *string
	if err := d.QueryRow(`SELECT deck_id, deck_name FROM seats WHERE seat = 0`).Scan(&deckID, &deckName); err != nil {
		t.Fatal(err)
	}
	if deckID == nil || *deckID != "d1" || deckName == nil || *deckName != "Atraxa" {
		t.Errorf("seat 0 deck = %v / %v, want d1 / Atraxa", deckID, deckName)
	}
	if err := d.QueryRow(`SELECT pending_discord_id FROM seats WHERE seat = 1`).Scan(&pending); err != nil || pending == nil || *pending != "123" {
		t.Errorf("seat 1 pending_discord_id = %v (%v)", pending, err)
	}
	// The new columns read NULL on every old row.
	var lastDeck, sourceURL *string
	if err := d.QueryRow(`SELECT last_deck FROM users WHERE id = 'u1'`).Scan(&lastDeck); err != nil || lastDeck != nil {
		t.Errorf("users.last_deck = %v (%v), want NULL", lastDeck, err)
	}
	if err := d.QueryRow(`SELECT source_url FROM decks WHERE id = 'd1'`).Scan(&sourceURL); err != nil || sourceURL != nil {
		t.Errorf("decks.source_url = %v (%v), want NULL", sourceURL, err)
	}
	assertRememberMeShape(t, d)

	// Deleting the deck the seat used nulls deck_id (the FK action) and
	// keeps the seat and its label.
	mustExec(t, d, `DELETE FROM decks WHERE id = 'd1'`)
	if err := d.QueryRow(`SELECT deck_id, deck_name FROM seats WHERE seat = 0`).Scan(&deckID, &deckName); err != nil {
		t.Fatal(err)
	}
	if deckID != nil || deckName == nil || *deckName != "Atraxa" {
		t.Errorf("after deleting the deck: deck_id = %v, deck_name = %v; want NULL and Atraxa", deckID, deckName)
	}
	assertForeignKeysOnEveryConn(t, d)
}

// A seat pointing at a deck that does not exist must still stop the
// migration, as it did for 0005: the runner's foreign_key_check is what
// keeps the rebuild honest.
func TestMigration0008WithDanglingDeckIDRollsBack(t *testing.T) {
	d := openAtVersion(t, 7)
	mustExec(t, d, `PRAGMA foreign_keys = OFF`)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'FNM', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, deck_id) VALUES ('g1', 0, 'p1', 'no-such-deck')`)
	mustExec(t, d, `PRAGMA foreign_keys = ON`)

	err := migrate(context.Background(), d)
	if err == nil || !strings.Contains(err.Error(), "dangling foreign key") {
		t.Fatalf("migrate: got %v, want a dangling-foreign-key refusal", err)
	}
	v, err := currentSchemaVersion(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	if v != 7 {
		t.Errorf("schema version after a refused migration = %d, want 7", v)
	}
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name IN ('user_settings', 'table_setups')`).Scan(&n); err != nil || n != 0 {
		t.Errorf("a refused migration left %d of its tables behind (%v)", n, err)
	}
}

func TestMigration0008SettingsConstraints(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	ins := `INSERT INTO user_settings (user_id, version, revision, body, updated_at) VALUES (?, 1, 1, ?, 1)`

	mustExec(t, d, ins, "u1", `{"a":1}`)
	if _, err := d.Exec(ins, "ghost", `{}`); err == nil {
		t.Error("user_settings.user_id accepted a user that does not exist")
	}
	mustExec(t, d, `DELETE FROM user_settings`)
	for name, body := range map[string]string{
		"not json":   `{nope`,
		"an array":   `[1]`,
		"a scalar":   `3`,
		"over 32KiB": `{"a":"` + strings.Repeat("x", 32768) + `"}`,
	} {
		if _, err := d.Exec(ins, "u1", body); err == nil {
			t.Errorf("user_settings accepted %s", name)
		}
	}
	// Exactly the cap is allowed.
	exact := `{"a":"` + strings.Repeat("x", 32768-8) + `"}`
	if len(exact) != 32768 {
		t.Fatalf("test body is %d bytes", len(exact))
	}
	mustExec(t, d, ins, "u1", exact)
}

// table_setups.game_id is deliberately not a foreign key.
func TestMigration0008SetupSurvivesItsGame(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'G', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO table_setups (user_id, body, game_id, updated_at) VALUES ('u1', '{}', 'g1', 1)`)
	mustExec(t, d, `DELETE FROM games WHERE id = 'g1'`)
	if n := countRows(t, d, "table_setups"); n != 1 {
		t.Errorf("table_setups has %d rows after its game was deleted, want 1", n)
	}
	if refs := referencesOf(t, d, "table_setups"); refs["user_id"] != "users" || len(refs) != 1 {
		t.Errorf("table_setups FKs = %v, want only user_id -> users", refs)
	}
}

// An older binary refuses a v8 database, as it does for any schema it
// does not know (ADR 0051), and the hand rollback documented in
// docs/environments.md leaves a database that rolls forward cleanly.
func TestMigration0008RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 0)

	// An older binary refuses a v8 database outright, by the generic
	// rule TestOpenSchemaTooNewRefuses pins (current > maxKnown), so the
	// rollback is a hand job: the additive pieces are undone and the
	// version row removed. The seats rebuild is harmless to an older
	// binary and is not undone.
	if v, err := currentSchemaVersion(ctx, d); err != nil || v < 8 {
		t.Fatalf("version = %d (%v), want >= 8", v, err)
	}
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, last_deck) VALUES ('u1', 'A', 1, 1, '{"kind":"prebuilt","id":"x"}')`)
	for _, q := range []string{
		`DROP TABLE user_settings`,
		`DROP TABLE table_setups`,
		`ALTER TABLE users DROP COLUMN last_deck`,
		`ALTER TABLE decks DROP COLUMN source_url`,
		`DELETE FROM schema_migrations WHERE version = 8`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 7 {
		t.Fatalf("after the rollback the version is %d, want 7", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	assertRememberMeShape(t, d)
	if n := countRows(t, d, "users"); n != 1 {
		t.Errorf("users: %d rows after the round trip, want 1", n)
	}
}

// assertRememberMeShape checks what 0008 promises and that the rebuild
// of seats left everything 0003 and 0005 gave it.
func assertRememberMeShape(t *testing.T, d *sql.DB) {
	t.Helper()
	if refs := referencesOf(t, d, "user_settings"); refs["user_id"] != "users" || len(refs) != 1 {
		t.Errorf("user_settings FKs = %v, want user_id -> users", refs)
	}
	refs := referencesOf(t, d, "seats")
	if refs["deck_id"] != "decks" || refs["user_id"] != "users" || refs["game_id"] != "games" {
		t.Errorf("seats FKs = %v", refs)
	}
	var onDelete string
	if err := d.QueryRow(`SELECT on_delete FROM pragma_foreign_key_list('seats') WHERE "from" = 'deck_id'`).Scan(&onDelete); err != nil || onDelete != "SET NULL" {
		t.Errorf("seats.deck_id ON DELETE = %q (%v), want SET NULL", onDelete, err)
	}
	if err := d.QueryRow(`SELECT on_delete FROM pragma_foreign_key_list('seats') WHERE "from" = 'game_id'`).Scan(&onDelete); err != nil || onDelete != "CASCADE" {
		t.Errorf("seats.game_id ON DELETE = %q (%v), want CASCADE", onDelete, err)
	}
	if got := strings.Join(indexesOn(t, d, "seats"), ","); got != "seats_pending_discord_id,seats_user_id" {
		t.Errorf("seats indexes = %s", got)
	}
	var leftovers int
	if err := d.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name LIKE '%\_new' ESCAPE '\'`).Scan(&leftovers); err != nil || leftovers != 0 {
		t.Errorf("%d *_new objects left in the schema (%v)", leftovers, err)
	}
	for _, c := range [][2]string{{"users", "last_deck"}, {"decks", "source_url"}} {
		var n int
		if err := d.QueryRow(`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, c[0], c[1]).Scan(&n); err != nil || n != 1 {
			t.Errorf("%s.%s missing (%v)", c[0], c[1], err)
		}
	}
	assertForeignKeysOnEveryConn(t, d)
}
