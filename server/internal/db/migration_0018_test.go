package db

import (
	"context"
	"encoding/json"
	"testing"
)

// Tests for migration 0018 (#3071): the pre-built deck ID izzet-aggro
// becomes mono-red-aggro in users.last_deck and in every bot of a
// remembered table setup, and nothing else changes.

func TestMigration0018RenamesTheDeckInLastDeckRecords(t *testing.T) {
	d := openAtVersion(t, 17)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, last_deck) VALUES
		('u1', 'A', 1, 1, '{"kind":"prebuilt","id":"izzet-aggro"}'),
		('u2', 'B', 1, 1, '{"kind":"prebuilt","id":"simic-ramp"}'),
		('u3', 'C', 1, 1, '{"kind":"library","id":"izzet-aggro"}'),
		('u4', 'D', 1, 1, NULL)`)

	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v17 -> latest: %v", err)
	}
	want := map[string]string{
		"u1": `{"kind":"prebuilt","id":"mono-red-aggro"}`,
		"u2": `{"kind":"prebuilt","id":"simic-ramp"}`,
		// A library deck's id is a decks row, not a pre-built ID.
		"u3": `{"kind":"library","id":"izzet-aggro"}`,
	}
	for id, w := range want {
		var got string
		if err := d.QueryRow(`SELECT last_deck FROM users WHERE id = ?`, id).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if !sameJSON(t, got, w) {
			t.Errorf("%s last_deck = %s, want %s", id, got, w)
		}
	}
	var null *string
	if err := d.QueryRow(`SELECT last_deck FROM users WHERE id = 'u4'`).Scan(&null); err != nil {
		t.Fatal(err)
	}
	if null != nil {
		t.Errorf("u4 last_deck = %q, want NULL", *null)
	}
}

func TestMigration0018RenamesTheDeckInTableSetups(t *testing.T) {
	d := openAtVersion(t, 17)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1), ('u2', 'B', 1, 1)`)
	before := `{"settings":{"starting_life":30},"bots":[` +
		`{"tier":"heuristic","deck_id":"izzet-aggro","name":"Bot izzet-aggro"},` +
		`{"tier":"random","deck_id":"simic-ramp","name":"Bot 2"},` +
		`{"tier":"strong","deck_id":"izzet-aggro","name":"Bot 3"}],` +
		`"tablemates":["00000000-0000-0000-0000-000000000001"]}`
	untouched := `{"settings":{},"bots":[{"tier":"random","deck_id":"esper-control","name":"Bot 1"}],"tablemates":[]}`
	mustExec(t, d, `INSERT INTO table_setups (user_id, body, game_id, updated_at) VALUES ('u1', ?, 'g1', 5), ('u2', ?, NULL, 6)`, before, untouched)

	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v17 -> latest: %v", err)
	}
	var got, gotUntouched string
	if err := d.QueryRow(`SELECT body FROM table_setups WHERE user_id = 'u1'`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	want := `{"settings":{"starting_life":30},"bots":[` +
		`{"tier":"heuristic","deck_id":"mono-red-aggro","name":"Bot izzet-aggro"},` +
		`{"tier":"random","deck_id":"simic-ramp","name":"Bot 2"},` +
		`{"tier":"strong","deck_id":"mono-red-aggro","name":"Bot 3"}],` +
		`"tablemates":["00000000-0000-0000-0000-000000000001"]}`
	if !sameJSON(t, got, want) {
		t.Errorf("setup body =\n %s\nwant\n %s", got, want)
	}
	if err := d.QueryRow(`SELECT body FROM table_setups WHERE user_id = 'u2'`).Scan(&gotUntouched); err != nil {
		t.Fatal(err)
	}
	if gotUntouched != untouched {
		t.Errorf("a setup with no izzet-aggro bot changed:\n %s\nwant\n %s", gotUntouched, untouched)
	}
	var gameID string
	var updated int64
	if err := d.QueryRow(`SELECT game_id, updated_at FROM table_setups WHERE user_id = 'u1'`).Scan(&gameID, &updated); err != nil {
		t.Fatal(err)
	}
	if gameID != "g1" || updated != 5 {
		t.Errorf("game_id, updated_at = %q, %d; want g1, 5", gameID, updated)
	}
}

// sameJSON compares two JSON documents by value, so key order and
// spacing do not matter but array order does.
func sameJSON(t *testing.T, a, b string) bool {
	t.Helper()
	var x, y any
	if err := json.Unmarshal([]byte(a), &x); err != nil {
		t.Fatalf("not JSON: %s: %v", a, err)
	}
	if err := json.Unmarshal([]byte(b), &y); err != nil {
		t.Fatalf("not JSON: %s: %v", b, err)
	}
	xb, _ := json.Marshal(x)
	yb, _ := json.Marshal(y)
	return string(xb) == string(yb)
}
