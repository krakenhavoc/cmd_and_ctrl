package testsm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Another creature entering pings each opponent for 1.
func TestMoltenGatekeeperPingsOnAnotherCreatureEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Molten Gatekeeper", "Artifact Creature — Golem", effectstest.MoltenGatekeeperOracle, false)
	lifeBefore := opp.Life

	effectstest.PushVanillaCreature(g, me.ID, "Fodder Bear", 1, 1)
	// pushVanillaCreature doesn't fire EventZoneMove, so simulate the
	// entry event the trigger watches for.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID})
	})
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != lifeBefore-1 {
		t.Errorf("opponent life %d -> %d, want -1", lifeBefore, opp.Life)
	}
}

// The Gatekeeper itself entering (its own unearthed return) does not
// ping anybody — "another" excludes it.
func TestMoltenGatekeeperUnearth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	id := effectstest.SeedGraveyardCard(t, g, "Molten Gatekeeper", "Artifact Creature — Golem", effectstest.MoltenGatekeeperOracle)
	lifeBefore := opp.Life

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Unearth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatalf("the unearthed creature should be on the battlefield")
	}
	if opp.Life != lifeBefore {
		t.Errorf("opponent life %d -> %d, the Gatekeeper's own return should not ping", lifeBefore, opp.Life)
	}
}
