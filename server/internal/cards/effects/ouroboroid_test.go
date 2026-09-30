package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ouroboroidOracle = "50d6fd91-23d3-4d32-804f-6233e4386904"

// TestOuroboroidPutsItsOwnPowerInCountersOnEachCreatureAtCombat — a
// 1/3 base Ouroboroid puts one counter on each creature the
// controller controls (itself included) at the beginning of combat.
func TestOuroboroidPutsItsOwnPowerInCountersOnEachCreatureAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ouro := pushCatalogPermanent(g, me.ID, "Ouroboroid", "Creature — Plant Wurm", ouroboroidOracle, false)
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)

	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	var ouroC, mineC, theirsC int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			switch c.InstanceID {
			case ouro:
				ouroC = c.Counters[game.CounterPlusOne]
			case mine:
				mineC = c.Counters[game.CounterPlusOne]
			case theirs:
				theirsC = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if ouroC != 1 {
		t.Errorf("Ouroboroid's own counters = %d, want 1 (its base power is 1)", ouroC)
	}
	if mineC != 1 {
		t.Errorf("controller's other creature counters = %d, want 1", mineC)
	}
	if theirsC != 0 {
		t.Errorf("opponent's creature counters = %d, want 0 (not \"you control\")", theirsC)
	}
}
