package testsex

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestImpulseExiledRiftCanBeCastForOverload(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.SeedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")
	me, rift := effectstest.ImpulseExile(t, g, "Cyclonic Rift", "Instant", effectstest.CyclonicRiftOracle)
	effectstest.AdvanceToMainForTest(t, g)

	// The offer is on the table for the caster.
	var offered bool
	g.ReadSnapshot(func() {
		card, _ := g.LookupCardForEffect(rift)
		grant := g.CastPermissionForLocked(me.ID, card, game.ZoneExile)
		for _, o := range g.CastOffersForLocked(me.ID, card, game.ZoneExile, grant) {
			if o != nil && o.Key == "overload" {
				offered = true
			}
		}
	})
	if !offered {
		t.Fatal("overload is not among the offers for a Rift exiled with a play permission (#2767)")
	}

	if err := g.CastSpell(me.ID, rift, game.CastSpellParams{FromZone: "exile", AlternativeCost: "overload"}); err != nil {
		t.Fatalf("CastSpell overload from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the overloaded Rift left the opponent's permanent on the battlefield")
	}
}

// The offers bound to another zone stay bound to it: flashback is
// claimable from the graveyard only, whatever opened exile.
func TestFlashbackIsStillRefusedFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.ImpulseExile(t, g, "Faithless Looting", "Sorcery", effectstest.FaithlessLootingOracle)
	effectstest.AdvanceToMainForTest(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", AlternativeCost: "flashback"}); err == nil {
		t.Fatal("flashback was claimed from exile")
	}
}
