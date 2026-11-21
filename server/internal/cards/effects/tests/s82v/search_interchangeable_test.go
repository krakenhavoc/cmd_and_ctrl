package testss82v

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// An agent seat cracking Evolving Wilds was offered one "take Swamp" move
// per Swamp in its library. Copies of a card answer a search the same
// way, so the seat is offered one move per kind of card.
func TestASearchOffersOneMovePerKindOfCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wilds := effectstest.PushCatalogPermanent(g, me.ID, "Evolving Wilds", "Land", effectstest.EvolvingWildsOracle, false)
	for i := 0; i < 6; i++ {
		effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
		effectstest.StapleLibraryCard(me, "Swamp", "Basic Land — Swamp")
	}
	if err := g.ActivateCatalogAbility(me.ID, wilds, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	var takes []string
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindChoice && !m.AlwaysLegal {
			takes = append(takes, m.Label)
		}
	}
	if len(takes) != 2 {
		t.Fatalf("offered %d picks, want one Forest and one Swamp: %q", len(takes), takes)
	}

	// Each offered pick is still accepted.
	effectstest.TakeSearchPrompts(t, g)
	lands := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && (c.Name == "Forest" || c.Name == "Swamp") {
			lands++
		}
	}
	if lands != 1 {
		t.Errorf("%d basics fetched, want 1", lands)
	}
}
