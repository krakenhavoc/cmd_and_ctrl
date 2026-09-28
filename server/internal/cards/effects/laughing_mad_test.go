package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const laughingMadOracle = "4f65e6a9-0d90-44b9-9b76-00814db2dbd8"

func TestLaughingMadDiscardsThenDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	before := me.Library.Size()

	_, paid := castWithDiscard(t, g, "Laughing Mad", laughingMadOracle, "Sorcery", 1)
	if !me.Graveyard.Contains(paid[0]) {
		t.Fatalf("the discard should be paid at announce, not on resolution")
	}
	passPriorityAroundTable(t, g)

	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if me.Hand.Size() != 2 {
		t.Errorf("hand = %d, want 2 (fodder discarded, spell cast, two drawn)", me.Hand.Size())
	}
}

// "Flashback for its flashback cost AND ANY ADDITIONAL COSTS": the
// discard is still owed on the graveyard cast path.
func TestLaughingMadFlashbackStillPaysTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	discard := handCard(me, "Fodder", "Sorcery")
	id := seedGraveyardCard(t, g, "Laughing Mad", "Instant", laughingMadOracle)
	before := me.Library.Size()

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		DiscardIDs: []uuid.UUID{discard},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	if !me.Graveyard.Contains(discard) {
		t.Fatalf("the discard should be paid at announce")
	}
	passPriorityAroundTable(t, g)

	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back instant should be exiled")
	}
}
