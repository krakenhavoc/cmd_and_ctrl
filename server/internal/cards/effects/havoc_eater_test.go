package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const havocEaterOracle = "769c511c-37ce-4a9d-85f8-1ba91c2daf82"

func TestHavocEaterIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(havocEaterOracle)
	if !ok || spec.Name != "Havoc Eater" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestHavocEaterGoadsOneCreaturePerOpponentAndGrowsByTheirTotalPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a, b, c := g.Seats[1], g.Seats[2], g.Seats[3]
	bigA := tdCreature(g, a.ID, "Big A", 4, 4)
	tdCreature(g, a.ID, "Small A", 1, 1)
	bigB := tdCreature(g, b.ID, "Big B", 3, 3)
	mine := tdCreature(g, me.ID, "Mine", 6, 6)
	_ = c // the third opponent controls no creature: "up to one"

	eater := rfdCast(t, g, "Havoc Eater", "Creature — Elemental", havocEaterOracle, 3, 3)
	if pick := latestPickTarget(g, me.ID); pick != nil && hasID(pick.PickTargetCards, mine) {
		t.Error("my own creature is offered")
	}
	pickCard(t, g, me.ID, bigA)
	pickCard(t, g, me.ID, bigB)
	passPriorityAroundTable(t, g)

	ca, _ := g.LookupCardForEffect(bigA)
	cb, _ := g.LookupCardForEffect(bigB)
	if !ca.IsGoadedBy(me.ID) || !cb.IsGoadedBy(me.ID) {
		t.Errorf("goaded: A=%v B=%v", ca.IsGoadedBy(me.ID), cb.IsGoadedBy(me.ID))
	}
	cm, _ := g.LookupCardForEffect(mine)
	if cm.IsGoaded() {
		t.Error("my own creature was goaded")
	}
	if n := countersOn(g, eater, game.CounterPlusOne); n != 7 {
		t.Errorf("+1/+1 counters = %d, want 7 (4 + 3)", n)
	}
}

func TestHavocEaterWithNoTargetsAddsNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	eater := rfdCast(t, g, "Havoc Eater", "Creature — Elemental", havocEaterOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, eater, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d, want 0", n)
	}
}
