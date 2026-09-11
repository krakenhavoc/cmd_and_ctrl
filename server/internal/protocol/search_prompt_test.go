package protocol

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestSearchPromptIsPrivateToTheSearcher guards the CR 400.2 half of
// the S22 search chooser. The searcher's prompt lists the matching
// library cards; every other seat must see neither their identities
// NOR their number, because both are facts about a hidden zone.
//
// Redaction alone is not enough here and that is the point of the
// test: an option list projected to backs still has a length, which
// would tell the table how many Forests are left in a library
// mid-fetch.
func TestSearchPromptIsPrivateToTheSearcher(t *testing.T) {
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 10)
		for j := range deck {
			deck[j] = game.NewCard("filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer %d: %v", i, err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 4))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	searcher, opponent := g.Seats[0], g.Seats[1]

	searcher.Library.Cards = nil
	for _, name := range []string{"Forest", "Island", "Mountain"} {
		searcher.Library.PushTop(game.Card{
			InstanceID: uuid.New(),
			Name:       name,
			TypeLine:   "Basic Land — " + name,
			Owner:      searcher.ID,
			Controller: searcher.ID,
		})
	}

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(searcher.ID,
			func(c game.Card) bool { return c.IsLand() },
			game.ZoneBattlefield, 1, false, true)
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("expected the search prompt to be queued, got %d choices", len(g.PendingChoices))
	}

	mine := FilterViewFor(ViewOfGame(g), searcher.ID.String())
	if len(mine.PendingChoices) != 1 {
		t.Fatalf("searcher sees %d choices, want 1", len(mine.PendingChoices))
	}
	own := mine.PendingChoices[0]
	if own.Kind != "search_library" {
		t.Fatalf("kind = %q, want search_library", own.Kind)
	}
	if len(own.Options) != 3 {
		t.Errorf("searcher sees %d candidates, want all 3", len(own.Options))
	}
	if own.SearchMax != 1 {
		t.Errorf("search_max = %d, want 1", own.SearchMax)
	}
	for _, c := range own.Options {
		if !c.KnownByYou || c.Name == "" {
			t.Errorf("searcher cannot identify their own candidate %s", c.InstanceID)
		}
	}

	theirs := FilterViewFor(ViewOfGame(g), opponent.ID.String())
	if len(theirs.PendingChoices) != 1 {
		t.Fatalf("opponent sees %d choices, want 1 (the prompt exists, just empty)", len(theirs.PendingChoices))
	}
	if n := len(theirs.PendingChoices[0].Options); n != 0 {
		t.Errorf("opponent sees %d search candidates, want 0 — the COUNT leaks the library", n)
	}
	if theirs.PendingChoices[0].SearchMax != 0 {
		t.Error("opponent sees the pick limit, which leaks what the search is for")
	}

	// Spectators and admins are held to the same rule.
	spectator := FilterViewFor(ViewOfGame(g), "")
	if n := len(spectator.PendingChoices[0].Options); n != 0 {
		t.Errorf("a spectator sees %d search candidates, want 0", n)
	}
}
