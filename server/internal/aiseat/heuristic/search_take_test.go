package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #3053: "find nothing" is on every tutor prompt's menu now, against a
// single match too, and the enumerator lists it FIRST (it is the choice's
// always-legal answer), so a tie goes to it. A bot must take a card it
// values at nothing rather than decline: declining needs a reason, and
// "I do not know what this is worth" is not one.
func TestASearchTakeBeatsFindingNothingEvenForACardWorthNothing(t *testing.T) {
	const choiceID = "choice-1"
	// The take names a card the view does not list, so the bot prices it
	// at nothing: the worst case for the empty answer's tie.
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withChoice(protocol.PendingChoiceView{
			ID: choiceID, Kind: "search_library", Chooser: seatID(0).String(), Reason: "Search",
		}))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "find nothing", map[string]any{"card_ids": []string{}}),
		choiceMove(t, 0, choiceID, "take the unknown card", map[string]any{"card_ids": []string{cardID(9)}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take the unknown card" {
		t.Fatalf("chose %q: the bot declined a search with a card on offer", got)
	}
}
