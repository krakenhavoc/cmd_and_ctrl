package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// random_target_test.go — #4051. A clause "chosen at random" on an
// activated ability is drawn by the engine as the ability is activated:
// the enumerator names no target for it, and every move it offers is
// one the engine accepts. The view marks the clause so the client's
// walk skips it the same way.

const oracleGoblinTestPilot = "4760bb87-56e9-47eb-b5ec-6fa4dd19156e"

func TestGoblinTestPilotEnumeratesNoTarget(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	pilot := battlefieldCard(g, active, game.Card{Name: "Goblin Test Pilot", TypeLine: "Creature — Goblin Pilot Wizard",
		OracleID: oracleGoblinTestPilot, Power: 0, Toughness: 2})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := activationsOf(legal.EnumerateFor(g, active.ID), pilot)
	if len(moves) == 0 {
		t.Fatal("Goblin Test Pilot's ability is not offered")
	}
	for _, m := range moves {
		if ids := boundTargetIDs(t, m); len(ids) != 0 {
			t.Fatalf("%q names %v; a target chosen at random is the engine's to draw", m.Label, ids)
		}
	}
	dispatchAll(t, g, active.ID, moves)

	v := protocol.ViewOfGameFor(g, active.ID.String())
	marked := false
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != pilot.String() {
			continue
		}
		for _, ab := range c.ActivatedAbilities {
			if ab.LegalTargets != nil && ab.LegalTargets.AtRandom {
				marked = true
			}
		}
	}
	if !marked {
		t.Fatal("the view does not mark the random clause")
	}
}
