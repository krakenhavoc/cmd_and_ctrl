package game

import "testing"

// energy_paid_or_lost_test.go — ADR 0129 §6: the per-turn tally of
// energy paid or lost (PlayerTurnTally.EnergyPaidOrLost). Every negative
// energy delta that lands counts: a payment through payEnergyLocked
// (CR 107.14) and a removal by an effect. Getting energy does not, a
// removal counts only what came off, and the tally resets with the turn.

func TestEnergyPaidOrLostCountsPaymentsAndRemovals(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushEnergySource(g, me, AbilityCost{Energy: 3})
	setEnergyForTest(t, g, me, 5)
	if got := g.EnergyPaidOrLostThisTurn(me.ID); got != 0 {
		t.Fatalf("getting energy counted %d, want 0", got)
	}

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := g.EnergyPaidOrLostThisTurn(me.ID); got != 3 {
		t.Fatalf("after paying three: %d, want 3", got)
	}

	// An effect removing five from a player with two takes two off, and
	// two is what counts.
	g.mu.Lock()
	err := g.AddPlayerCounterForEffect(me.ID, CounterEnergy, -5)
	g.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if got := g.EnergyPaidOrLostThisTurn(me.ID); got != 5 {
		t.Errorf("after losing the last two: %d, want 5", got)
	}
	if got := g.EnergyPaidOrLostThisTurn(opp.ID); got != 0 {
		t.Errorf("the opponent's tally is %d, want 0", got)
	}
	// Other player counters are not energy.
	g.mu.Lock()
	err = g.AddPlayerCounterForEffect(me.ID, CounterPoison, 2)
	if err == nil {
		err = g.AddPlayerCounterForEffect(me.ID, CounterPoison, -2)
	}
	g.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if got := g.EnergyPaidOrLostThisTurn(me.ID); got != 5 {
		t.Errorf("poison changed the energy tally to %d", got)
	}

	// It rides the undo clone and the snapshot with the rest of the
	// turn's record.
	if got := g.Clone().EnergyPaidOrLostThisTurn(me.ID); got != 5 {
		t.Errorf("clone: %d, want 5", got)
	}

	passTurn(t, g)
	if got := g.EnergyPaidOrLostThisTurn(me.ID); got != 0 {
		t.Errorf("a new turn kept %d, want 0", got)
	}
}
