package heuristic_test

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// hand_discard_cost_test.go — #1600: legal.MoveCost.Hand reaching the
// policy. "Discard your hand" names no card in the move's params (the
// engine refuses ids for it), so without the count Slate of Ancestry's
// activation would read as free and a bot would throw away seven cards
// to draw two.

func slateView(id string, controller int) protocol.CardView {
	return protocol.CardView{
		InstanceID: id, Name: "Slate of Ancestry", TypeLine: "Artifact",
		Owner: seatID(controller).String(), Controller: seatID(controller).String(), KnownByYou: true,
	}
}

const slateLabel = "Slate of Ancestry: {4}, {T}, Discard your hand: Draw a card for each creature you control."

// The same activation ranks lower the more cards it throws away, and an
// empty hand is priced as nothing.
func TestHandDiscardCostIsPriced(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(slateView(cardID(1), 0), land(cardID(10), 0)))
	empty := slateLabel + " (empty hand)"
	full := slateLabel + " (seven cards)"
	in := input(0, v,
		passMove(0),
		activateMove(t, 0, cardID(1), empty, nil),
		activateMove(t, 0, cardID(1), full, &legal.MoveCost{Hand: 7}),
	)
	ranked := heuristic.New().Rank(context.Background(), in)
	value := map[int]float64{}
	for _, c := range ranked {
		value[c.Index] = c.Value
	}
	if value[2] >= value[1] {
		t.Errorf("discarding seven cards is valued %.2f, the empty-hand activation %.2f — want it lower", value[2], value[1])
	}
}
