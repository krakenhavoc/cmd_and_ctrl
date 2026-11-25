package game

import (
	"testing"

	"github.com/google/uuid"
)

// untap_counter_removal_test.go — #6095: "If this creature would untap
// during your untap step, remove a +1/+1 counter from it instead. If you
// do, untap it." (Bewitching Leechcraft's granted ability). The catalog
// half, where an Aura grants it, is in cards/effects.

const leechOracle = "untap-counter-removal-probe"

func stubLeechDef(t *testing.T, instances int) {
	t.Helper()
	removals := make([]UntapCounterRemoval, instances)
	for i := range removals {
		removals[i] = UntapCounterRemoval{Counter: CounterPlusOne, Label: "probe"}
	}
	stubRowDefs(t, map[string]*CardDef{leechOracle: {UntapCounterRemovals: removals}})
}

func leechCreature(g *Game, controller uuid.UUID, counters int) uuid.UUID {
	id := pushTappedPermanent(g, controller, "Leeched Bear", leechOracle, "Creature — Bear", true)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.Power, c.Toughness = 2, 2
		if counters > 0 {
			c.Counters = map[string]int{CounterPlusOne: counters}
		}
	})
	return id
}

func leechState(g *Game, id uuid.UUID) (tapped bool, counters int) {
	g.ReadSnapshot(func() {
		c := findBattlefieldCard(g, id)
		tapped, counters = c.Tapped, c.Counters[CounterPlusOne]
	})
	return
}

func TestUntapCounterRemovalUntapsOnlyByPayingACounter(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	stubLeechDef(t, 1)
	me := g.Seats[0]
	paid := leechCreature(g, me.ID, 2)
	broke := leechCreature(g, me.ID, 0)
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	advanceToStepOfSeat(t, g, 0, StepUpkeep)
	if tapped, n := leechState(g, paid); tapped || n != 1 {
		t.Errorf("with two counters: tapped %v, %d counters; want untapped with 1", tapped, n)
	}
	if tapped, _ := leechState(g, broke); !tapped {
		t.Error("with no counter the creature untapped during its controller's untap step")
	}
	// Outside the untap step the untap is not replaced.
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	g.WithWriteLock(func() {
		if err := g.UntapTargetForEffect(broke); err != nil {
			t.Fatalf("UntapTargetForEffect: %v", err)
		}
	})
	if tapped, _ := leechState(g, broke); tapped {
		t.Error("an untap effect outside the untap step was replaced")
	}
}

// Another player's untap step is not "your untap step": a creature untapped
// there by an untap-step permission keeps its counters.
func TestUntapCounterRemovalOnlyDuringItsControllersUntapStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	stubLeechDef(t, 1)
	opp := g.Seats[1]
	id := leechCreature(g, opp.ID, 1)
	advanceToStepOfSeat(t, g, 2, StepUpkeep)
	if tapped, n := leechState(g, id); tapped || n != 0 {
		t.Fatalf("after its controller's untap step: tapped %v, %d counters; want untapped with 0", tapped, n)
	}
}

// Two instances each replace the untap the other leaves: the creature
// untaps only if both find a counter, and each that found one removed it.
func TestUntapCounterRemovalTwoInstancesNeedTwoCounters(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	stubLeechDef(t, 2)
	me := g.Seats[0]
	one := leechCreature(g, me.ID, 1)
	two := leechCreature(g, me.ID, 3)
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	advanceToStepOfSeat(t, g, 0, StepUpkeep)
	if tapped, n := leechState(g, one); !tapped || n != 0 {
		t.Errorf("with one counter: tapped %v, %d counters; want tapped with 0", tapped, n)
	}
	if tapped, n := leechState(g, two); tapped || n != 1 {
		t.Errorf("with three counters: tapped %v, %d counters; want untapped with 1", tapped, n)
	}
}
