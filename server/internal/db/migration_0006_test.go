package db

import (
	"context"
	"strings"
	"testing"
)

// Tests for migration 0006 (deck requests, ADR 0095 §3): two new
// tables and the rate-limit index, added to a database an older binary
// left at v5 without touching anything already in it.

func TestMigration0006FromEmpty(t *testing.T) {
	d := openAtVersion(t, 0)
	for _, table := range []string{"deck_requests", "deck_request_asks"} {
		if n := countRows(t, d, table); n != 0 {
			t.Errorf("%s: %d rows in a fresh database", table, n)
		}
	}
	if got := strings.Join(indexesOn(t, d, "deck_request_asks"), ","); got != "deck_request_asks_requester_at" {
		t.Errorf("deck_request_asks indexes = %s", got)
	}
	// Neither table references users: the bot files for Discord
	// members who may never have signed in.
	for _, table := range []string{"deck_requests", "deck_request_asks"} {
		if refs := referencesOf(t, d, table); len(refs) != 0 {
			t.Errorf("%s FKs = %v, want none", table, refs)
		}
	}
}

func TestMigration0006PreservesAPopulatedV5Database(t *testing.T) {
	d := openAtVersion(t, 5)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'Alice', 1, 1)`)
	mustExec(t, d, `INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
		VALUES ('d1', 'u1', 'Atraxa', 'text', '1 Atraxa', '[]', 1, 1, 1)`)

	if err := migrateTo(context.Background(), d, 6); err != nil {
		t.Fatalf("migrateTo(6): %v", err)
	}
	for table, want := range map[string]int{"users": 1, "decks": 1, "deck_requests": 0, "deck_request_asks": 0} {
		if got := countRows(t, d, table); got != want {
			t.Errorf("%s: %d rows, want %d", table, got, want)
		}
	}

	mustExec(t, d, `INSERT INTO deck_requests (deck_key, issue_number, issue_url, created_at)
		VALUES ('moxfield:abc', 12, 'https://github.com/o/r/issues/12', 1700000000000)`)
	if _, err := d.Exec(`INSERT INTO deck_requests (deck_key, issue_number, issue_url, created_at)
		VALUES ('moxfield:abc', 13, 'https://github.com/o/r/issues/13', 1)`); err == nil {
		t.Error("deck_requests accepted a second row for one deck_key")
	}
	mustExec(t, d, `INSERT INTO deck_request_asks (deck_key, requester, at) VALUES
		('moxfield:abc', 'discord:1', 1), ('moxfield:abc', 'discord:1', 2)`)
	if n := countRows(t, d, "deck_request_asks"); n != 2 {
		t.Errorf("deck_request_asks: %d rows, want 2 (one requester may ask twice)", n)
	}
	assertForeignKeysOnEveryConn(t, d)
}
