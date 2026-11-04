package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// blight_cost_test.go — #4684: a "Blight N" cost (Gristle Glutton's "{T},
// Blight 1:") puts -1/-1 counters on a creature the bot chose, and the
// policy has to see the price. The enumerator offers one move per
// creature; legal.MoveCost.Blight carries N and the creature, priced by
// blightCounterCost.

func blightActivateMove(t *testing.T, seat int, src, label, creature string, n int) legal.Move {
	t.Helper()
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: label, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{
			"source_card_id": src, "ability_index": 0, "blight_ids": []string{creature},
		}),
		Cost: &legal.MoveCost{Blight: &legal.CounterPrice{CardID: uuid.MustParse(creature), Counter: "-1/-1", N: n}},
	}
}

// Two activations that differ only in the blight: the policy takes the
// one that pays nothing.
func TestABlightCostIsPricedOnAnActivation(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			land(cardID(1), 0),
			land(cardID(2), 0),
			creature(cardID(30), 0, "Test Bear", 3, 3),
		))
	const costly, free = "Blights The Bear: loot", "Blights Nothing: the same ability"
	in := input(0, v, passMove(0), blightActivateMove(t, 0, cardID(1), costly, cardID(30), 1),
		exileActivateMove(t, 0, cardID(2), free))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != free {
		t.Errorf("the bot chose %q — a blight cost should price below paying nothing", got)
	}
}

// Of two creatures it could blight, the bot puts the counter on the wall
// that shrugs it off, not on the 1/1 it would kill, whichever is offered
// first.
func TestABlightCostGoesOnTheCreatureItCostsLeast(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			land(cardID(1), 0),
			creature(cardID(30), 0, "Test Goblin", 1, 1),
			creature(cardID(31), 0, "Test Wall", 0, 6),
		))
	const kills, shrugs = "Blight the Goblin", "Blight the Wall"
	in := input(0, v, passMove(0),
		blightActivateMove(t, 0, cardID(1), kills, cardID(30), 1),
		blightActivateMove(t, 0, cardID(1), shrugs, cardID(31), 1))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == kills {
		t.Errorf("the bot chose %q — a blight that kills a creature should price below one that costs a wall a point of toughness", got)
	}
}
