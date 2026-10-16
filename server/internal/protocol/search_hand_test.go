package protocol

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3988: a "search that player's graveyard, hand, and library" prompt
// marks its hand candidates (search_hand) and says every zone is searched
// (search_all_zones). The searcher sees it; the searched player and the
// rest of the table are told none of it.
func TestSearchPromptMarksTheHandHalf(t *testing.T) {
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
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	searcher, victim := g.Seats[0], g.Seats[1]
	handID := uuid.New()
	victim.Hand.PushTop(game.Card{InstanceID: handID, Name: "Lightning Bolt", TypeLine: "Instant", Owner: victim.ID, Controller: victim.ID})
	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: searcher.ID, LibraryOwner: victim.ID, Pred: func(c game.Card) bool { return c.Name == "Lightning Bolt" },
			Dest: game.ZoneExile, Unbounded: true, Shuffle: true, AlsoGraveyardAndHand: true,
		})
	})

	mine := FilterViewFor(ViewOfGame(g), searcher.ID.String())
	if len(mine.PendingChoices) != 1 {
		t.Fatalf("searcher sees %d choices, want 1", len(mine.PendingChoices))
	}
	pc := mine.PendingChoices[0]
	if len(pc.Options) != 1 || pc.Options[0].InstanceID != handID.String() {
		t.Fatalf("options = %+v, want the Bolt in their hand", pc.Options)
	}
	if len(pc.SearchHand) != 1 || pc.SearchHand[0] != handID.String() || !pc.SearchAllZones {
		t.Errorf("search_hand = %v, search_all_zones = %v", pc.SearchHand, pc.SearchAllZones)
	}

	theirs := FilterViewFor(ViewOfGame(g), victim.ID.String())
	if c := theirs.PendingChoices[0]; len(c.SearchHand) != 0 || c.SearchAllZones || len(c.Options) != 0 {
		t.Errorf("the searched player sees the search: %+v", c)
	}
}
