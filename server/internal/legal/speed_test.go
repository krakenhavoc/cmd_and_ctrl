package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// speed_test.go — a "Max speed —" activated ability (ADR 0138, #2122)
// in the agreement style: the enumerator offers it only when the
// engine would accept it, which is only at max speed. Perilous Snare's
// "Max speed — {T}: Put a +1/+1 counter on target creature or Vehicle
// you control. Activate only as a sorcery."

const oraclePerilousSnare = "cff7499a-44b0-4a2a-b96c-ed2cafb1a90d"

func TestMaxSpeedActivationIsOfferedOnlyAtMaxSpeed(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	snare := battlefieldCard(g, active, game.Card{
		Name: "Perilous Snare", TypeLine: "Artifact", OracleID: oraclePerilousSnare,
	})
	battlefieldCard(g, active, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceTo(t, g, game.StepPrecombatMain)

	g.SetSpeedForTest(active.ID, 3)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), snare); len(acts) != 0 {
		t.Fatalf("offered below max speed: %v", labels(acts))
	}

	g.SetSpeedForTest(active.ID, game.MaxSpeed)
	acts := activationsOf(legal.EnumerateFor(g, active.ID), snare)
	if len(acts) == 0 {
		t.Fatal("not offered at max speed in the sorcery window")
	}
	dispatchAll(t, g, active.ID, acts)
}
