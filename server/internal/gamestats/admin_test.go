package gamestats

import (
	"context"
	"fmt"
	"testing"
)

func TestOverviewAggregatesAFixture(t *testing.T) {
	s, d := openStore(t)
	for _, u := range []string{"u-me", "u-other"} {
		if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, u, "name-"+u); err != nil {
			t.Fatal(err)
		}
	}
	seedSeat(t, d, 1000, "A", "u-me", "d1", "D", "win", 1, 7, `["Atraxa"]`)
	seedSeat(t, d, 2000, "B", "u-me", "d1", "D", "loss", 3, 9, `["Atraxa"]`)
	seedSeat(t, d, 3000, "C", "u-me", "", "", "draw", nil, 4, `["Tymna","Thrasios"]`)
	seedSeat(t, d, 4000, "D", "u-me", "", "", "none", nil, 0, `[]`)
	seedSeat(t, d, 5000, "E", "u-other", "", "", "win", 1, 5, `["Secret Commander"]`)
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := d.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	// One replay record with an incomplete log, one ended game with no
	// record, and one still running (neither counts as unrecorded).
	exec(`UPDATE game_stats SET source = 'replay', log_complete = 0 WHERE game_id = (SELECT id FROM games WHERE name = 'B')`)
	exec(`INSERT INTO games (id, name, state, created_at, ended_at) VALUES ('g-unrecorded', 'U', 'ended', 1, 6000)`)
	exec(`INSERT INTO games (id, name, state, created_at) VALUES ('g-running', 'R', 'active', 1)`)
	// Cards: Sol Ring cast in two games, Bolt in one.
	exec(`INSERT INTO seat_card_stats VALUES ((SELECT id FROM games WHERE name = 'A'), 0, 'Sol Ring', NULL, 3, 1)`)
	exec(`INSERT INTO seat_card_stats VALUES ((SELECT id FROM games WHERE name = 'B'), 0, 'Sol Ring', NULL, 2, 1)`)
	exec(`INSERT INTO seat_card_stats VALUES ((SELECT id FROM games WHERE name = 'A'), 0, 'Lightning Bolt', NULL, 1, 2)`)

	got, err := s.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := (Totals{Games: 5, Live: 4, Replay: 1, LogComplete: 4, LogIncomplete: 1, Unrecorded: 1}); got.Totals != want {
		t.Errorf("totals = %+v, want %+v", got.Totals, want)
	}
	if len(got.Recent) != 5 || got.Recent[0].Name != "E" || got.Recent[4].Name != "A" || got.Recent[1].Turns != 0 || got.Recent[3].Turns != 9 {
		t.Errorf("recent = %+v", got.Recent)
	}
	// Five unnamed seats (the losers) share one guest label; D's own seat
	// are not counted.
	if len(got.People) != 3 {
		t.Fatalf("people = %+v", got.People)
	}
	guest, me, other := got.People[0], got.People[1], got.People[2]
	if guest.Kind != "guest" || guest.Games != 5 || guest.Losses != 5 || guest.AvgPlace == nil || *guest.AvgPlace != 2 {
		t.Errorf("guest = %+v", guest)
	}
	if me.Name != "name-u-me" || me.Kind != "account" || me.Games != 3 || me.Wins != 1 || me.Losses != 1 || me.Draws != 1 || me.AvgPlace == nil || *me.AvgPlace != 2 {
		t.Errorf("me = %+v", me)
	}
	if other.Name != "name-u-other" || other.Games != 1 || other.Wins != 1 {
		t.Errorf("other = %+v", other)
	}
	if len(got.CommandersGames) != 4 || got.CommandersGames[0].Name != "Other Person's Commander" || got.CommandersGames[0].Games != 5 ||
		got.CommandersGames[1].Name != "Atraxa" || got.CommandersGames[1].Games != 2 {
		t.Errorf("commanders by games = %+v", got.CommandersGames)
	}
	if got.CommandersWins[0].Name != "Atraxa" || got.CommandersWins[0].Wins != 1 || got.CommandersWins[1].Name != "Secret Commander" {
		t.Errorf("commanders by wins = %+v", got.CommandersWins)
	}
	if len(got.Cards) != 2 || got.Cards[0] != (CardCasts{Name: "Sol Ring", Casts: 5, Games: 2}) || got.Cards[1].Name != "Lightning Bolt" {
		t.Errorf("cards = %+v", got.Cards)
	}
}

func TestOverviewEmptyAndCapped(t *testing.T) {
	s, d := openStore(t)
	got, err := s.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Totals != (Totals{}) || got.Recent == nil || got.People == nil || got.CommandersGames == nil || got.CommandersWins == nil || got.Cards == nil {
		t.Errorf("empty overview = %+v, want zero totals and empty lists", got)
	}
	for i := 0; i < AdminListMax+5; i++ {
		seedSeat(t, d, int64(1000+i), fmt.Sprintf("G%d", i), "", "", "", "win", 1, 3, fmt.Sprintf(`["Cmdr %02d"]`, i))
	}
	got, err = s.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Recent) != AdminListMax || len(got.CommandersGames) != AdminListMax || len(got.CommandersWins) != AdminListMax {
		t.Errorf("lists = %d/%d/%d, want %d", len(got.Recent), len(got.CommandersGames), len(got.CommandersWins), AdminListMax)
	}
}

// ADR 0165 §9: deleted accounts fold into one row, agent seats are their
// own kind, and bot and agent seats stay out of the commander and card
// lists.
func TestOverviewKeepsDeletedBotsAndAgentsOutOfHumanStats(t *testing.T) {
	s, d := openStore(t)
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := d.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	for _, u := range []string{"u-gone1", "u-gone2"} {
		exec(`INSERT INTO users (id, display_name, created_at, last_seen_at, deleted_at) VALUES (?, 'Deleted player', 1, 1, 5)`, u)
	}
	exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u-live', 'Live', 1, 1)`)
	seedSeat(t, d, 1000, "A", "u-gone1", "", "", "win", 1, 5, `["Gone One Cmdr"]`)
	seedSeat(t, d, 2000, "B", "u-gone2", "", "", "loss", 2, 5, `["Gone Two Cmdr"]`)
	seedSeat(t, d, 3000, "C", "u-live", "", "", "win", 1, 5, `["Human Cmdr"]`)
	seedSeat(t, d, 4000, "D", "", "", "", "win", 1, 5, `["Agent Cmdr"]`)
	seedSeat(t, d, 5000, "E", "", "", "", "win", 1, 5, `["Bot Cmdr"]`)
	exec(`UPDATE seats SET agent_client = 'claude-code', guest_name = 'Agent Seat' WHERE seat = 0 AND game_id = (SELECT id FROM games WHERE name = 'D')`)
	exec(`UPDATE seats SET bot_tier = 'heuristic' WHERE seat = 0 AND game_id = (SELECT id FROM games WHERE name = 'E')`)
	exec(`INSERT INTO seat_card_stats VALUES ((SELECT id FROM games WHERE name = 'C'), 0, 'Human Card', NULL, 2, 1)`)
	exec(`INSERT INTO seat_card_stats VALUES ((SELECT id FROM games WHERE name = 'D'), 0, 'Agent Card', NULL, 9, 1)`)
	exec(`INSERT INTO seat_card_stats VALUES ((SELECT id FROM games WHERE name = 'E'), 0, 'Bot Card', NULL, 9, 1)`)

	got, err := s.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	byKind := map[string][]PersonRecord{}
	for _, p := range got.People {
		byKind[p.Kind] = append(byKind[p.Kind], p)
		if p.Name == "Deleted player" || p.Name == "Agent Seat" {
			t.Errorf("person row %+v leaks a tombstone or seat name", p)
		}
	}
	if d := byKind["deleted"]; len(d) != 1 || d[0].Name != "Deleted players" || d[0].Games != 2 || d[0].Wins != 1 || d[0].Losses != 1 {
		t.Errorf("deleted = %+v, want one row of both accounts", d)
	}
	if a := byKind["agent"]; len(a) != 1 || a[0].Name != "AI agent (claude-code)" || a[0].Games != 1 || a[0].Wins != 1 {
		t.Errorf("agent = %+v", a)
	}
	if b := byKind["bot"]; len(b) != 1 || b[0].Name != "Bot (heuristic)" {
		t.Errorf("bot = %+v", b)
	}
	if l := byKind["account"]; len(l) != 1 || l[0].Name != "Live" {
		t.Errorf("account = %+v", l)
	}
	for _, g := range append(append([]Group{}, got.CommandersGames...), got.CommandersWins...) {
		if g.Name == "Agent Cmdr" || g.Name == "Bot Cmdr" {
			t.Errorf("commander %q is from a bot or agent seat", g.Name)
		}
	}
	found := false
	for _, g := range got.CommandersGames {
		found = found || g.Name == "Human Cmdr"
	}
	if !found {
		t.Errorf("human commander missing: %+v", got.CommandersGames)
	}
	if len(got.Cards) != 1 || got.Cards[0].Name != "Human Card" {
		t.Errorf("cards = %+v, want only the human seat's", got.Cards)
	}
}
