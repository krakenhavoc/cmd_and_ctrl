package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4262 (CR 101.4, CR 616.1): each player's exile starts from the previous
// player's, so two libraries that each ask for an ordering open ONE prompt at
// a time, the active player's first.
func TestDoomsdayExcruciatorExilesEachLibraryInTurnWithOnePromptAtATime(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	other := g.Seats[2]
	advanceToMain(t, g)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("active seat %d, want 0", g.Turn.ActiveSeat)
	}
	// A card near the top of each library, each with an ordering prompt.
	promptCard := func(p *game.Player) uuid.UUID {
		var id uuid.UUID
		g.WithWriteLock(func() { id = p.Library.Cards[len(p.Library.Cards)-3].InstanceID })
		orderExileOf(g, id)
		return id
	}
	myCard, otherCard := promptCard(me), promptCard(other)

	castFromHandForTest(t, g, me, "Doomsday Excruciator", "Creature — Demon", "", s58DoomsdayExcruciator, game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	// First prompt: the active player's library. The other library has not
	// been touched yet.
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open, want exactly the active player's", len(g.PendingChoices))
	}
	if !other.Library.Contains(otherCard) || other.Library.Size() <= 6 {
		t.Fatalf("the second library was exiled from (size %d) while the first prompt was open", other.Library.Size())
	}
	if !me.Library.Contains(myCard) {
		t.Fatal("the first prompt's card left its library before the answer")
	}
	answerExileOrder(t, g)

	// The second prompt opens only now, for the next library.
	if me.Library.Size() != 6 {
		t.Errorf("first library keeps %d cards after its answer, want 6", me.Library.Size())
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open after the first answer, want the second library's", len(g.PendingChoices))
	}
	if !other.Library.Contains(otherCard) {
		t.Fatal("the second prompt's card left its library before the answer")
	}
	answerExileOrder(t, g)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts still open", len(g.PendingChoices))
	}
	for i, p := range g.Seats {
		if p.Library.Size() != 6 {
			t.Errorf("seat %d keeps %d cards, want 6", i, p.Library.Size())
		}
	}
	for _, id := range []uuid.UUID{myCard, otherCard} {
		if c, ok := findInZone(g.Exile, id); !ok || !c.FaceDown {
			t.Errorf("prompted card %s is not in exile face down", id)
		}
	}
}
