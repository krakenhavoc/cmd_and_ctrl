package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_b_dash_test.go — dash (CR 702.109), #3172, on Ragavan.

func castRagavan(t *testing.T, g *game.Game, me *game.Player, alt, pool string) uuid.UUID {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Ragavan, Nimble Pilferer", "Legendary Creature — Monkey Pirate", "{R}", effectstest.RagavanOracle, []string{"R"})
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 2, 1
			}
		}
	})
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: alt}); err != nil {
		t.Fatalf("cast Ragavan (%q): %v", alt, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// Dash {1}{R}: Ragavan arrives with haste, attacks the turn it is cast,
// and goes back to its owner's hand at the beginning of the next end
// step.
func TestRagavanDashHasteThenReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := castRagavan(t, g, me, "dash", "CR")
	if !effectstest.OnBattlefieldByID(g, id) {
		t.Fatal("dashed Ragavan should be on the battlefield")
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Fatal("dashed Ragavan has haste")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("a dashed creature attacks the turn it was cast: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, id) || !me.Hand.Contains(id) {
		t.Errorf("Ragavan should be back in its owner's hand at the end step (battlefield=%v hand=%v)",
			effectstest.OnBattlefieldByID(g, id), me.Hand.Contains(id))
	}
}

// Cast normally for {R}: summoning sick, and it stays.
func TestRagavanCastNormallyHasNoHasteAndStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := castRagavan(t, g, me, "", "R")
	if effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Fatal("a normal cast gives no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err == nil {
		t.Error("summoning sick Ragavan must not attack")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, id) {
		t.Error("a normally cast Ragavan stays on the battlefield")
	}
}

// A dashed creature that has already died is not returned from the
// graveyard.
func TestRagavanDashReturnSkipsADeadCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := castRagavan(t, g, me, "dash", "CR")
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}, game.DestroyOptions{}) })
	if !me.Graveyard.Contains(id) {
		t.Fatal("setup: Ragavan should be dead")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("a dead dashed creature stays in the graveyard")
	}
}
