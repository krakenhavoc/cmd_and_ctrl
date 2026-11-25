package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// life_auction_test.go — #6212: a bot asked to top the high bid of a life
// auction passes, which is what lets the bidding end.
func TestABotBidderPasses(t *testing.T) {
	pol := heuristic.New()
	const id = "55555555-5555-5555-5555-555555555555"
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withChoice(protocol.PendingChoiceView{
		ID: id, Kind: "pay_amount", Chooser: seatID(1).String(),
		PayAmount: &protocol.PayAmountView{Min: 4, Max: 1000000, Unit: "bid", Resource: "bid", NoMax: true},
	}))
	in := input(1, v,
		choiceMove(t, 1, id, "pass", map[string]any{"amount": 0}),
		choiceMove(t, 1, id, "top", map[string]any{"amount": 4}))
	pass, top := rankValue(t, pol, in, "pass"), rankValue(t, pol, in, "top")
	if pass <= top {
		t.Errorf("pass %.3f, top %.3f; a bot bidder must pass", pass, top)
	}
}
