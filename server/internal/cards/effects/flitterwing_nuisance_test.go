package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const flitterwingNuisanceOracle = "9148f2cb-5594-463c-a21b-a67706ec46d1"

func TestFlitterwingNuisanceIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(flitterwingNuisanceOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// It enters with a -1/-1 counter; paying {2}{U} and removing it makes a
// this-turn trigger that draws for each creature you control dealing
// combat damage to a player or a planeswalker, and for nothing else.
func TestFlitterwingNuisanceDrawsOnCombatDamageThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	nuisance := rfdCast(t, g, "Flitterwing Nuisance", "Creature — Faerie Rogue", flitterwingNuisanceOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, nuisance, game.CounterMinusOne); got != 1 {
		t.Fatalf("entered with %d -1/-1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, nuisance, "flying") {
		t.Error("no flying")
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ox := b12Creature(g, me.ID, "My Ox", "Creature — Ox", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := b12Permanent(g, opp.ID, "Their Walker", "Legendary Planeswalker — Test")

	floatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, nuisance, 0, game.ActivateAbilityParams{Strict: true, CounterKind: game.CounterMinusOne}); err == nil {
		t.Fatal("activated without {2}{U} in the pool")
	}
	floatForTest(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, nuisance, 0, game.ActivateAbilityParams{Strict: true, CounterKind: game.CounterMinusOne}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, nuisance, game.CounterMinusOne); got != 0 {
		t.Fatalf("the -1/-1 counter did not pay the cost: %d left", got)
	}

	hand, theirHand := me.Hand.Size(), opp.Hand.Size()
	b27Damage(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	dealCombatDamageToPlayer(g, theirs, me.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || opp.Hand.Size() != theirHand {
		t.Fatalf("noncombat damage or an opponent's creature drew: hand %d → %d, theirs %d → %d",
			hand, me.Hand.Size(), theirHand, opp.Hand.Size())
	}
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	dealCombatDamageToPlayer(g, ox, walker, 3)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("drew %d for a player and a planeswalker hit, want 2", me.Hand.Size()-hand)
	}

	floatForTest(g, me, "UUU")
	if err := g.ActivateCatalogAbility(me.ID, nuisance, 0, game.ActivateAbilityParams{Strict: true, CounterKind: game.CounterMinusOne}); err == nil {
		t.Error("activated with no counter to remove")
	}

	advanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	hand = me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("the trigger outlasted the turn: hand %d → %d", hand, me.Hand.Size())
	}
}
