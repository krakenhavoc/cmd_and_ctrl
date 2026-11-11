package db

import (
	"context"
	"testing"
)

// Tests for migration 0036 (the format of a saved deck, ADR 0222 §2):
// one decks column defaulting to 'commander', so every deck saved before
// formats existed reads as a Commander deck, and the hand rollback in
// docs/environments.md.

// deckFormatVersion is this migration's number.
const deckFormatVersion = 36

const insertDeckBeforeFormats = `INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
	VALUES ('d1', 'u1', 'Old deck', 'text', '1 Sol Ring', '["Atraxa"]', 100, 1, 1)`

func TestMigration0036ReadsExistingDecksAsCommander(t *testing.T) {
	d := openAtVersion(t, deckFormatVersion-1)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, insertDeckBeforeFormats)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate to latest: %v", err)
	}
	var format, source string
	if err := d.QueryRow(`SELECT format, source_text FROM decks WHERE id = 'd1'`).Scan(&format, &source); err != nil {
		t.Fatal(err)
	}
	if format != "commander" {
		t.Errorf("a deck saved before formats reads as %q; want commander", format)
	}
	if source != "1 Sol Ring" {
		t.Errorf("the saved list changed: %q", source)
	}
	// A new row that names no format is a Commander deck too.
	mustExec(t, d, `INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
		VALUES ('d2', 'u1', 'New deck', 'text', '1 Sol Ring', '[]', 100, 2, 2)`)
	if err := d.QueryRow(`SELECT format FROM decks WHERE id = 'd2'`).Scan(&format); err != nil {
		t.Fatal(err)
	}
	if format != "commander" {
		t.Errorf("a new deck with no format reads as %q; want commander", format)
	}
	mustExec(t, d, `UPDATE decks SET format = 'standard' WHERE id = 'd2'`)
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0036RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, deckFormatVersion)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, insertDeckBeforeFormats)
	mustExec(t, d, `UPDATE decks SET format = 'standard' WHERE id = 'd1'`)
	mustExec(t, d, `ALTER TABLE decks DROP COLUMN format`)
	mustExec(t, d, `DELETE FROM schema_migrations WHERE version = ?`, deckFormatVersion)
	if v, _ := currentSchemaVersion(ctx, d); v != deckFormatVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, deckFormatVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("roll forward again: %v", err)
	}
	var format string
	if err := d.QueryRow(`SELECT format FROM decks WHERE id = 'd1'`).Scan(&format); err != nil {
		t.Fatal(err)
	}
	if format != "commander" {
		t.Errorf("rolled forward: format %q; want commander, the default every deck gets back", format)
	}
}
