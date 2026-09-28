package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dreadReturnOracle = "352b64d2-2ae5-44ee-a64f-94932ef545d3"

// The hand-cast half: an ordinary reanimation.
func TestDreadReturnReanimatesFromHand(t *testing.T) {
	g := newCatalogGame(t)
	victim := seedGraveyardCard(t, g, "Fodder Bear", "Creature — Bear", "")

	castCatalogSpell(t, g, "Dread Return", "Sorcery", dreadReturnOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim},
	})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(victim) {
		t.Fatalf("the targeted creature should be on the battlefield")
	}
}

// Declared caveat: the flashback cost has no sacrifice-cost component
// in AlternativeCost, so the card offers no graveyard cast path at
// all.
func TestDreadReturnHasNoFlashbackOffer(t *testing.T) {
	if game.CardCastableFromZone(dreadReturnOracle, game.ZoneGraveyard) {
		t.Fatalf("Dread Return should not be castable from the graveyard — its flashback cost isn't implemented")
	}
	if offers := game.AlternativeCostsOfferedFromZone(dreadReturnOracle, game.ZoneGraveyard); len(offers) != 0 {
		t.Fatalf("Dread Return should offer no alternative cost from the graveyard, got %+v", offers)
	}
}
