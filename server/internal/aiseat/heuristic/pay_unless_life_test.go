package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// pay_unless_life_test.go — ADR 0131 §2 and §5: a pay_unless "pay" that
// spends 2 life on a symbol (a ward {B} under K'rrik) is priced as the
// Phyrexian life it is, so it takes the same floor of 10 a cast does.

func wardLifeInput(t *testing.T, life int) aiseat.Input {
	t.Helper()
	const choiceID = "choice-ward"
	v := newView([]protocol.PlayerView{newSeat(0, withLife(life)), newSeat(1)},
		withChoice(protocol.PendingChoiceView{
			ID:               choiceID,
			Kind:             "pay_unless",
			Chooser:          seatID(0).String(),
			Reason:           "Ward {B}",
			PayCost:          "{B}",
			PhyrexianSymbols: 1,
			PhyrexianGranted: 1,
		}))
	yes, no := true, false
	pay := choiceMove(t, 0, choiceID, "pay 2 life", map[string]any{"apply": yes, "phyrexian_life": 1})
	pay.Cost = &legal.MoveCost{Life: 2, PhyrexianLife: 2}
	return input(0, v, choiceMove(t, 0, choiceID, "decline", map[string]any{"apply": no}), pay)
}

func TestHeuristicPaysAWardWithLifeAboveTheFloor(t *testing.T) {
	for _, tc := range []struct {
		life int
		want string
	}{
		{20, "pay 2 life"},
		{12, "pay 2 life"}, // lands on 10, the floor itself
		{11, "decline"},
		{5, "decline"},
	} {
		in := wardLifeInput(t, tc.life)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
			t.Errorf("at %d life the bot chose %q, want %q", tc.life, got, tc.want)
		}
	}
}
