package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// instant_loyalty_answers_test.go — a planeswalker that opens its own
// loyalty abilities at instant speed is offered them while an
// opponent's spell waits on the stack, so smart autopass reads what they
// answer. Teferi, Master of Time's +1 is a loot: it declares value and
// does not stop its controller on every opponent's spell. Before the
// declaration the row was undeclared, and an undeclared row interacts
// (ADR 0142 owner answer 3): a bot arena replay counted 57 such stops
// in eight games.

const oracleTeferiMasterOfTimeIL = "e802fb53-7cf5-46bc-8a0b-f99cf5c20f74"

func TestTeferiMasterOfTimesLootDoesNotInteract(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	teferi := battlefieldCard(g, seat, game.Card{
		Name: "Teferi, Master of Time", TypeLine: "Legendary Planeswalker — Teferi", ManaCost: "{2}{U}",
		OracleID: oracleTeferiMasterOfTimeIL, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	bear := battlefieldCard(g, seat, game.Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2,
	})
	targetedBy(g, opponentOf(g, seat), bear)

	var loot *legal.Move
	for _, m := range movesFrom(legal.EnumerateFor(g, seat.ID), teferi, legal.KindActivate) {
		if m.Label == "Teferi, Master of Time: +1: Draw a card, then discard a card." {
			loot = &m
		}
	}
	if loot == nil {
		t.Fatal("Teferi's +1 is not offered with an opponent's spell on the stack: the instant-speed timing is off")
	}
	if loot.Interacts || loot.CombatInteracts || loot.HasTargets {
		t.Errorf("Teferi's +1: interacts=%v combat_interacts=%v has_targets=%v, want none (a loot answers nothing)",
			loot.Interacts, loot.CombatInteracts, loot.HasTargets)
	}
}
