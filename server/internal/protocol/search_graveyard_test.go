package protocol

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3217: a "search your library and/or graveyard" prompt marks which of
// its options sit in the graveyard (search_graveyard) and leaves
// search_min absent ("and/or" lets the searcher look only in the library).
// The searcher sees it; the opponent is told none of it, like search_max.
func TestSearchPromptMarksTheGraveyardHalf(t *testing.T) {
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
	searcher, opponent := g.Seats[0], g.Seats[1]
	searcher.Library.Cards = nil
	libID, gyID := uuid.New(), uuid.New()
	searcher.Library.PushTop(game.Card{InstanceID: libID, Name: "Command Tower", TypeLine: "Land", Owner: searcher.ID, Controller: searcher.ID})
	searcher.Graveyard.PushTop(game.Card{InstanceID: gyID, Name: "Command Tower", TypeLine: "Land", Owner: searcher.ID, Controller: searcher.ID})
	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: searcher.ID, Pred: func(c game.Card) bool { return c.Name == "Command Tower" },
			Dest: game.ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
		})
	})

	mine := FilterViewFor(ViewOfGame(g), searcher.ID.String())
	if len(mine.PendingChoices) != 1 {
		t.Fatalf("searcher sees %d choices, want 1", len(mine.PendingChoices))
	}
	pc := mine.PendingChoices[0]
	if len(pc.Options) != 2 {
		t.Fatalf("options = %d, want both Towers", len(pc.Options))
	}
	if len(pc.SearchGraveyard) != 1 || pc.SearchGraveyard[0] != gyID.String() {
		t.Errorf("search_graveyard = %v, want just the graveyard Tower %s", pc.SearchGraveyard, gyID)
	}
	if pc.SearchMin != 0 {
		t.Errorf("search_min = %d, want 0: finding nothing stays legal", pc.SearchMin)
	}
	// #4516: the searcher chooses the zones, so the client is told to ask
	// whether the library was searched; the opponent is not told.
	if !pc.SearchLibraryOptional {
		t.Error("search_library_optional is absent from an \"and/or graveyard\" prompt")
	}

	theirs := FilterViewFor(ViewOfGame(g), opponent.ID.String())
	if theirs.PendingChoices[0].SearchLibraryOptional {
		t.Error("the opponent is told which kind of search it is")
	}
	if len(theirs.PendingChoices[0].SearchGraveyard) != 0 || len(theirs.PendingChoices[0].Options) != 0 {
		t.Errorf("the opponent sees the search: %+v", theirs.PendingChoices[0])
	}
}
