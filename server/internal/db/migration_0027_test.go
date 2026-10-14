package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0027 (a saved deck's estimated power level,
// #3671): old decks arrive unrated, the column refuses a value off the
// 1–5 scale, and the hand rollback in docs/environments.md.

const deckPowerVersion = 27

func TestMigration0027PreservesAPopulatedV26DatabaseWithUnratedDecks(t *testing.T) {
	d := openAtVersion(t, deckPowerVersion-1)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'Alice', 1, 1)`)
	mustExec(t, d, `INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
		VALUES ('d1', 'u1', 'Atraxa', 'text', '1 Atraxa', '["Atraxa"]', 100, 1, 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v26 -> latest: %v", err)
	}
	var level sql.NullFloat64
	if err := d.QueryRow(`SELECT power_level FROM decks WHERE id = 'd1'`).Scan(&level); err != nil {
		t.Fatal(err)
	}
	if level.Valid {
		t.Errorf("an existing deck: power_level %v; want NULL", level)
	}
	mustExec(t, d, `UPDATE decks SET power_level = 3.5 WHERE id = 'd1'`)
	for _, bad := range []string{"0.5", "5.5"} {
		if _, err := d.Exec(`UPDATE decks SET power_level = ` + bad + ` WHERE id = 'd1'`); err == nil {
			t.Errorf("power_level = %s was accepted; want the CHECK to refuse it", bad)
		}
	}
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0027RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, deckPowerVersion)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'Alice', 1, 1)`)
	mustExec(t, d, `INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at, power_level)
		VALUES ('d1', 'u1', 'Atraxa', 'text', '1 Atraxa', '["Atraxa"]', 100, 1, 1, 4)`)
	for _, q := range []string{
		`ALTER TABLE decks DROP COLUMN power_level`,
		`DELETE FROM schema_migrations WHERE version = 27`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != deckPowerVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, deckPowerVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	var level sql.NullFloat64
	if err := d.QueryRow(`SELECT power_level FROM decks WHERE id = 'd1'`).Scan(&level); err != nil {
		t.Fatal(err)
	}
	if level.Valid {
		t.Errorf("after the round trip the deck is still rated: %v", level)
	}
}
