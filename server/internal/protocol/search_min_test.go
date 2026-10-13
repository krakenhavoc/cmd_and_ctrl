package protocol

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3053: the wire carries the floor. A search for a stated quality
// leaves search_min absent (zero: "Find nothing" is legal, CR 701.23b);
// an unconditional one sends 1 (CR 701.23d) and the client drops the
// button. The opponent is told neither, like search_max.
func TestSearchPromptCarriesTheFloor(t *testing.T) {
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
	for _, name := range []string{"Forest", "Island"} {
		searcher.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Basic Land — " + name,
			Owner: searcher.ID, Controller: searcher.ID,
		})
	}

	for _, tc := range []struct {
		name string
		pred func(game.Card) bool
		want int
	}{
		{"stated quality, one match", func(c game.Card) bool { return c.Name == "Forest" }, 0},
		{"unconditional", nil, 1},
	} {
		g.WithWriteLock(func() {
			g.PendingChoices = nil
			_ = g.SearchLibraryForEffect(searcher.ID, tc.pred, game.ZoneHand, 1, false, true)
		})
		mine := FilterViewFor(ViewOfGame(g), searcher.ID.String())
		if len(mine.PendingChoices) != 1 {
			t.Fatalf("%s: searcher sees %d choices, want 1", tc.name, len(mine.PendingChoices))
		}
		if got := mine.PendingChoices[0].SearchMin; got != tc.want {
			t.Errorf("%s: search_min = %d, want %d", tc.name, got, tc.want)
		}
		theirs := FilterViewFor(ViewOfGame(g), opponent.ID.String())
		if got := theirs.PendingChoices[0].SearchMin; got != 0 {
			t.Errorf("%s: the opponent sees search_min = %d, which leaks what the search is for", tc.name, got)
		}
	}
}
