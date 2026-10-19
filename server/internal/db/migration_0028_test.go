package db

import (
	"context"
	"testing"
)

// Tests for migration 0028 (the cards each deck request listed, #4135):
// a v27 database arrives with an empty table, a deck lists a card once,
// the table references no one, and the hand rollback in
// docs/environments.md.

const deckRequestCardsVersion = 28

func TestMigration0028PreservesAPopulatedV27DatabaseWithRequests(t *testing.T) {
	d := openAtVersion(t, deckRequestCardsVersion-1)
	mustExec(t, d, `INSERT INTO deck_requests (deck_key, issue_number, issue_url, created_at)
		VALUES ('moxfield:abc', 12, 'https://github.com/o/r/issues/12', 1)`)
	mustExec(t, d, `INSERT INTO deck_request_asks (deck_key, requester, at) VALUES ('moxfield:abc', 'discord:1', 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v27 -> latest: %v", err)
	}
	for table, want := range map[string]int{"deck_requests": 1, "deck_request_asks": 1, "deck_request_cards": 0} {
		if got := countRows(t, d, table); got != want {
			t.Errorf("%s: %d rows, want %d", table, got, want)
		}
	}
	mustExec(t, d, `INSERT INTO deck_request_cards (deck_key, card_key, card_name, issue_number)
		VALUES ('moxfield:abc', 'oracle-1', 'Sol Ring', 12)`)
	if _, err := d.Exec(`INSERT INTO deck_request_cards (deck_key, card_key, card_name, issue_number)
		VALUES ('moxfield:abc', 'oracle-1', 'Sol Ring', 13)`); err == nil {
		t.Error("deck_request_cards listed one card twice for one deck")
	}
	// The same card under another deck is its own row.
	mustExec(t, d, `INSERT INTO deck_request_cards (deck_key, card_key, card_name, issue_number)
		VALUES ('archidekt:9', 'oracle-1', 'Sol Ring', 14)`)
	if refs := referencesOf(t, d, "deck_request_cards"); len(refs) != 0 {
		t.Errorf("deck_request_cards FKs = %v, want none", refs)
	}
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0028RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, deckRequestCardsVersion)
	mustExec(t, d, `INSERT INTO deck_request_cards (deck_key, card_key, card_name, issue_number)
		VALUES ('moxfield:abc', 'oracle-1', 'Sol Ring', 12)`)
	for _, q := range []string{
		`DROP TABLE deck_request_cards`,
		`DELETE FROM schema_migrations WHERE version = 28`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != deckRequestCardsVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, deckRequestCardsVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	if n := countRows(t, d, "deck_request_cards"); n != 0 {
		t.Errorf("after the round trip the table has %d rows, want 0", n)
	}
}
