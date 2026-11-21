package testso

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// TestOnceEachTurnGateReopensForTheReturningPermanent — the bug: the
// gate stayed shut for the rest of the turn because TurnTally is
// keyed by instance ID and an instance ID survives a zone change. The
// permanent that comes back is a new object (CR 400.7) and gets its
// own first time.
func TestOnceEachTurnGateReopensForTheReturningPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	probe := effectstest.PushCatalogPermanent(g, me.ID, "Gate Probe", "Creature — Test", effectstest.GateProbeOracle, false)
	start := me.Life

	effectstest.EnterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start+1 {
		t.Fatalf("life %d after the first creature entered, want %d", me.Life, start+1)
	}

	effectstest.EnterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start+1 {
		t.Fatalf("life %d after a second creature entered, want %d — the gate is once each turn",
			me.Life, start+1)
	}

	effectstest.BounceAndReturnForTest(t, g, probe, me.ID)

	effectstest.EnterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start+2 {
		t.Errorf("life %d after the permanent left and came back, want %d — the returning "+
			"permanent is a new object and its once-each-turn clause may fire again (CR 400.7)",
			me.Life, start+2)
	}
}
