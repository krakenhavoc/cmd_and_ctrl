package testsex

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFlitterwingNuisanceIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.FlitterwingNuisanceOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// It enters with a -1/-1 counter; paying {2}{U} and removing it makes a
// this-turn trigger that draws for each creature you control dealing
// combat damage to a player or a planeswalker, and for nothing else.
func TestFlitterwingNuisanceDrawsOnCombatDamageThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	nuisance := effectstest.RfdCast(t, g, "Flitterwing Nuisance", "Creature — Faerie Rogue", effectstest.FlitterwingNuisanceOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, nuisance, game.CounterMinusOne); got != 1 {
		t.Fatalf("entered with %d -1/-1 counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, nuisance, "flying") {
		t.Error("no flying")
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, me.ID, "My Ox", "Creature — Ox", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := effectstest.B12Permanent(g, opp.ID, "Their Walker", "Legendary Planeswalker — Test")

	effectstest.FloatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, nuisance, 0, game.ActivateAbilityParams{Strict: true, CounterKind: game.CounterMinusOne}); err == nil {
		t.Fatal("activated without {2}{U} in the pool")
	}
	effectstest.FloatForTest(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, nuisance, 0, game.ActivateAbilityParams{Strict: true, CounterKind: game.CounterMinusOne}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, nuisance, game.CounterMinusOne); got != 0 {
		t.Fatalf("the -1/-1 counter did not pay the cost: %d left", got)
	}

	hand, theirHand := me.Hand.Size(), opp.Hand.Size()
	effectstest.B27Damage(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || opp.Hand.Size() != theirHand {
		t.Fatalf("noncombat damage or an opponent's creature drew: hand %d → %d, theirs %d → %d",
			hand, me.Hand.Size(), theirHand, opp.Hand.Size())
	}
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealCombatDamageToPlayer(g, ox, walker, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("drew %d for a player and a planeswalker hit, want 2", me.Hand.Size()-hand)
	}

	effectstest.FloatForTest(g, me, "UUU")
	if err := g.ActivateCatalogAbility(me.ID, nuisance, 0, game.ActivateAbilityParams{Strict: true, CounterKind: game.CounterMinusOne}); err == nil {
		t.Error("activated with no counter to remove")
	}

	effectstest.AdvanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	hand = me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("the trigger outlasted the turn: hand %d → %d", hand, me.Hand.Size())
	}
}
