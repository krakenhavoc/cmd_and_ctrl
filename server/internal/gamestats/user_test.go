package gamestats

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// seedSeat writes one game with one seat of stats and a seats row for
// userID ("" for a seat with no person).
func seedSeat(t *testing.T, d *db.DB, at int64, name, userID, deckID, deckName, result string, place any, turns int, cmdrs string) {
	t.Helper()
	id := uuid.NewString()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := d.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	exec(`INSERT INTO games (id, name, state, created_at, ended_at) VALUES (?, ?, 'ended', 1, ?)`, id, name, at)
	var uid any
	if userID != "" {
		uid = userID
	}
	var did any
	if deckID != "" {
		did = deckID
		exec(`INSERT OR IGNORE INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
		      VALUES (?, ?, ?, 'text', '', '[]', 100, 1, 1)`, deckID, userID, deckName)
	}
	exec(`INSERT INTO seats (game_id, seat, player_id, user_id, deck_id, deck_name) VALUES (?, 0, 'p0', ?, ?, ?)`, id, uid, did, deckName)
	// A second seat that is nobody's, so Players counts both.
	exec(`INSERT INTO seats (game_id, seat, player_id) VALUES (?, 1, 'p1')`, id)
	exec(`INSERT INTO game_stats VALUES (?, 1, 'live', 1, ?, 1, 0, ?)`, id, turns, at)
	for seat, r := range []string{result, "loss"} {
		p := place
		if seat == 1 {
			p = 2
		}
		c := cmdrs
		if seat == 1 {
			c = `["Other Person's Commander"]`
		}
		exec(`INSERT INTO seat_stats VALUES (?, ?, ?, ?, NULL, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ?)`, id, seat, r, p, c)
	}
}

func TestForUserRecordGroupsAndRecent(t *testing.T) {
	s, d := openStore(t)
	for _, u := range []string{"u-me", "u-other"} {
		if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, u, u); err != nil {
			t.Fatal(err)
		}
	}
	seedSeat(t, d, 1000, "A", "u-me", "d1", "Old Name", "win", 1, 7, `["Atraxa"]`)
	seedSeat(t, d, 2000, "B", "u-me", "d1", "Atraxa Superfriends", "loss", 3, 9, `["Atraxa"]`)
	seedSeat(t, d, 3000, "C", "u-me", "", "", "draw", nil, 4, `["Tymna","Thrasios"]`)
	seedSeat(t, d, 4000, "D", "u-me", "d2", "Closed", "none", nil, 0, `[]`)
	// Someone else's games must never reach u-me.
	seedSeat(t, d, 5000, "E", "u-other", "dx", "Theirs", "win", 1, 5, `["Secret Commander"]`)
	seedSeat(t, d, 6000, "F", "", "", "", "win", 1, 5, `["Bot Commander"]`)

	got, err := s.ForUser(context.Background(), "u-me")
	if err != nil {
		t.Fatal(err)
	}
	if got.Games != 3 || got.Wins != 1 || got.Losses != 1 || got.Draws != 1 || got.Unfinished != 1 {
		t.Errorf("record = %+v, want 3 games 1/1/1 and 1 unfinished", got.Tally)
	}
	// Places 1 and 3 over the two placed seats; the draw has none.
	if got.AvgPlace == nil || *got.AvgPlace != 2 {
		t.Errorf("avg place = %v, want 2", got.AvgPlace)
	}
	if len(got.ByDeck) != 1 || got.ByDeck[0].Name != "Atraxa Superfriends" || got.ByDeck[0].Games != 2 {
		t.Errorf("by deck = %+v, want one deck of 2 games under its newest name", got.ByDeck)
	}
	if len(got.ByCommander) != 2 || got.ByCommander[0].Name != "Atraxa" || got.ByCommander[0].Games != 2 ||
		got.ByCommander[1].Name != "Tymna / Thrasios" {
		t.Errorf("by commander = %+v", got.ByCommander)
	}
	if len(got.Recent) != 3 || got.Recent[0].Name != "C" || got.Recent[2].Name != "A" ||
		got.Recent[0].Players != 2 || got.Recent[1].Turns != 9 {
		t.Errorf("recent = %+v", got.Recent)
	}
}

// Another person's seats, and seats nobody holds, never appear.
func TestForUserNeverShowsAnotherPersonsSeats(t *testing.T) {
	s, d := openStore(t)
	for _, u := range []string{"u-me", "u-other"} {
		if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, u, u); err != nil {
			t.Fatal(err)
		}
	}
	seedSeat(t, d, 5000, "E", "u-other", "dx", "Theirs", "win", 1, 5, `["Secret Commander"]`)
	seedSeat(t, d, 6000, "F", "", "", "", "win", 1, 5, `["Bot Commander"]`)

	got, err := s.ForUser(context.Background(), "u-me")
	if err != nil {
		t.Fatal(err)
	}
	if got.Games != 0 || got.Unfinished != 0 || len(got.ByDeck) != 0 || len(got.ByCommander) != 0 || len(got.Recent) != 0 || got.AvgPlace != nil {
		t.Errorf("u-me has no seats but got %+v", got)
	}
	// The other seat of u-other's game (the loser, seat 1) is nobody's.
	for _, user := range []string{"", "nobody"} {
		got, err := s.ForUser(context.Background(), user)
		if err != nil || got.Games != 0 || len(got.Recent) != 0 {
			t.Errorf("ForUser(%q) = %+v, %v; want nothing", user, got, err)
		}
	}
	other, err := s.ForUser(context.Background(), "u-other")
	if err != nil || other.Games != 1 || other.Wins != 1 {
		t.Errorf("u-other = %+v, %v; want their own win", other, err)
	}
	if len(other.Recent) != 1 || other.Recent[0].Commanders[0] != "Secret Commander" {
		t.Errorf("u-other recent = %+v", other.Recent)
	}
}

// A bot or agent seat that carries the caller's user id is not a game
// they played themselves (ADR 0165 §9).
func TestForUserExcludesAgentAndBotSeats(t *testing.T) {
	s, d := openStore(t)
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u-me', 'me', 1, 1)`); err != nil {
		t.Fatal(err)
	}
	seedSeat(t, d, 1000, "Mine", "u-me", "", "", "win", 1, 5, `["Mine"]`)
	seedSeat(t, d, 2000, "Agent", "u-me", "", "", "win", 1, 5, `["Agent Cmdr"]`)
	seedSeat(t, d, 3000, "Bot", "u-me", "", "", "win", 1, 5, `["Bot Cmdr"]`)
	for _, q := range []string{
		`UPDATE seats SET agent_client = 'claude-code' WHERE seat = 0 AND game_id = (SELECT id FROM games WHERE name = 'Agent')`,
		`UPDATE seats SET bot_tier = 'heuristic' WHERE seat = 0 AND game_id = (SELECT id FROM games WHERE name = 'Bot')`,
	} {
		if _, err := d.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.ForUser(context.Background(), "u-me")
	if err != nil {
		t.Fatal(err)
	}
	if got.Games != 1 || len(got.Recent) != 1 || got.Recent[0].Name != "Mine" || len(got.ByCommander) != 1 {
		t.Errorf("got %+v, want only the caller's own game", got)
	}
}

// TopCards sums the caller's own human seats' casts, counted games only,
// most first, capped, and never another seat's, a bot's or an agent's.
func TestForUserTopCards(t *testing.T) {
	s, d := openStore(t)
	for _, u := range []string{"u-me", "u-other"} {
		if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, u, u); err != nil {
			t.Fatal(err)
		}
	}
	seedSeat(t, d, 1000, "A", "u-me", "", "", "win", 1, 5, `["X"]`)
	seedSeat(t, d, 2000, "B", "u-me", "", "", "loss", 2, 5, `["X"]`)
	seedSeat(t, d, 3000, "Closed", "u-me", "", "", "none", nil, 0, `[]`)
	seedSeat(t, d, 4000, "Agent", "u-me", "", "", "win", 1, 5, `["X"]`)
	seedSeat(t, d, 5000, "Theirs", "u-other", "", "", "win", 1, 5, `["X"]`)
	if _, err := d.Exec(`UPDATE seats SET agent_client = 'claude-code' WHERE seat = 0 AND game_id = (SELECT id FROM games WHERE name = 'Agent')`); err != nil {
		t.Fatal(err)
	}
	cast := func(game string, seat int, name string, n int) {
		t.Helper()
		if _, err := d.Exec(`INSERT INTO seat_card_stats VALUES ((SELECT id FROM games WHERE name = ?), ?, ?, NULL, ?, 1)`, game, seat, name, n); err != nil {
			t.Fatal(err)
		}
	}
	cast("A", 0, "Sol Ring", 2)
	cast("B", 0, "Sol Ring", 3)
	cast("A", 0, "Arcane Signet", 1)
	cast("A", 1, "Opponent Card", 9) // seat 1 is nobody's
	cast("Closed", 0, "Unfinished Card", 9)
	cast("Agent", 0, "Agent Card", 9)
	cast("Theirs", 0, "Other Card", 9)
	for i := 0; i < 12; i++ {
		cast("B", 0, "Filler "+string(rune('A'+i)), 1)
	}

	got, err := s.ForUser(context.Background(), "u-me")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.TopCards) != TopCardsMax {
		t.Fatalf("top cards = %d, want the cap %d: %+v", len(got.TopCards), TopCardsMax, got.TopCards)
	}
	if got.TopCards[0] != (TopCard{Name: "Sol Ring", Casts: 5}) {
		t.Errorf("first = %+v, want Sol Ring x5", got.TopCards[0])
	}
	for _, c := range got.TopCards {
		switch c.Name {
		case "Opponent Card", "Unfinished Card", "Agent Card", "Other Card":
			t.Errorf("top cards hold %q, which is not the caller's own counted human seat", c.Name)
		}
	}
	none, err := s.ForUser(context.Background(), "nobody")
	if err != nil || none.TopCards == nil || len(none.TopCards) != 0 {
		t.Errorf("empty = %+v, %v; want a non-nil empty list", none.TopCards, err)
	}
}
