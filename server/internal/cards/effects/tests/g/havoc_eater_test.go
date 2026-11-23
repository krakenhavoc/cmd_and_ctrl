package testsg

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const havocEaterOracle = "769c511c-37ce-4a9d-85f8-1ba91c2daf82"

func TestHavocEaterIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(havocEaterOracle)
	if !ok || spec.Name != "Havoc Eater" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestHavocEaterGoadsOneCreaturePerOpponentAndGrowsByTheirTotalPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a, b, c := g.Seats[1], g.Seats[2], g.Seats[3]
	bigA := effectstest.TdCreature(g, a.ID, "Big A", 4, 4)
	effectstest.TdCreature(g, a.ID, "Small A", 1, 1)
	bigB := effectstest.TdCreature(g, b.ID, "Big B", 3, 3)
	mine := effectstest.TdCreature(g, me.ID, "Mine", 6, 6)
	_ = c // the third opponent controls no creature: "up to one"

	eater := effectstest.RfdCast(t, g, "Havoc Eater", "Creature — Elemental", havocEaterOracle, 3, 3)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil && effectstest.HasID(pick.PickTargetCards, mine) {
		t.Error("my own creature is offered")
	}
	effectstest.PickCard(t, g, me.ID, bigA)
	effectstest.PickCard(t, g, me.ID, bigB)
	effectstest.PassPriorityAroundTable(t, g)

	ca, _ := g.LookupCardForEffect(bigA)
	cb, _ := g.LookupCardForEffect(bigB)
	if !ca.IsGoadedBy(me.ID) || !cb.IsGoadedBy(me.ID) {
		t.Errorf("goaded: A=%v B=%v", ca.IsGoadedBy(me.ID), cb.IsGoadedBy(me.ID))
	}
	cm, _ := g.LookupCardForEffect(mine)
	if cm.IsGoaded() {
		t.Error("my own creature was goaded")
	}
	if n := effectstest.CountersOn(g, eater, game.CounterPlusOne); n != 7 {
		t.Errorf("+1/+1 counters = %d, want 7 (4 + 3)", n)
	}
}

func TestHavocEaterWithNoTargetsAddsNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	eater := effectstest.RfdCast(t, g, "Havoc Eater", "Creature — Elemental", havocEaterOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, eater, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d, want 0", n)
	}
}
