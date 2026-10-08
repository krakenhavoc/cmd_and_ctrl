package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const oracleGildedGhoda = "5f4939dc-6e87-47fa-8287-9da60d4b5db2"

// Saddle (CR 702.171, #2695) is crew's cost over OTHER creatures. The
// enumerator has to leave the Mount out of its own payment, and offer
// nothing at all when the Mount is the only creature, or the engine
// refuses a move the bot was told was legal (#544).
func TestSaddleIsOfferedWithASetThatLeavesTheMountOut(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	mount := battlefieldCard(g, active, game.Card{
		Name: "Gilded Ghoda", TypeLine: "Creature — Horse Mount", ManaCost: "{1}{R}",
		Power: 2, Toughness: 2, OracleID: oracleGildedGhoda,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	// The Mount's own power would clear Saddle 1; it may not pay.
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), mount); len(acts) != 0 {
		t.Fatalf("a Mount with nothing else to tap was offered a saddle: %v", labels(acts))
	}

	battlefieldCard(g, active, creature("Saddler", "{1}{G}", 2, 2))
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, mount)
	if len(acts) != 1 {
		t.Fatalf("want one saddle activation, got %d: %v", len(acts), labels(acts))
	}
	var p struct {
		CrewIDs []string `json:"crew_ids"`
	}
	if err := json.Unmarshal(acts[0].Params, &p); err != nil {
		t.Fatalf("params: %v", err)
	}
	if len(p.CrewIDs) == 0 {
		t.Fatal("a saddle activation must name the creatures that pay for it")
	}
	for _, id := range p.CrewIDs {
		if id == mount.String() {
			t.Error("the saddle payment names the Mount itself")
		}
	}
	// Whatever is offered, the engine accepts.
	dispatchAll(t, g, active.ID, moves)

	// And at instant speed nothing is offered.
	advanceTo(t, g, game.StepDeclareAttackers)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), mount); len(acts) != 0 {
		t.Errorf("a saddle was offered in combat: %v", labels(acts))
	}
}
