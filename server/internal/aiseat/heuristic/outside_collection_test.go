package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// outside_collection_test.go — ADR 0198 §4: a pick from the card database
// brings the named card into the bot's hand, so the bot takes the best
// suggestion rather than nothing. The winning answer is listed last, so
// the test does not pass on the enumerator's order.

func collectionChoice(id, reason string, lo, hi int, suggestions ...protocol.CardView) protocol.PendingChoiceView {
	for i := range suggestions {
		suggestions[i].InstanceID = ""
	}
	return protocol.PendingChoiceView{
		ID: id, Kind: "outside_collection", Chooser: seatID(0).String(),
		FromPlayer: seatID(0).String(), Reason: reason,
		ChooseMin: lo, ChooseMax: hi, Count: hi, ChooseDestination: "outside",
		CollectionOptions: suggestions,
	}
}

func TestACollectionPickTakesTheBestSuggestion(t *testing.T) {
	const choiceID = "choice-collection"
	small := creature(cardID(1), 0, "Bear", 2, 2)
	big := creature(cardID(2), 0, "Giant", 6, 6)
	big.ManaCost = "{4}{R}{R}"
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(land(cardID(30), 0)),
		withChoice(collectionChoice(choiceID, "Living Wish — reveal a creature or land card you own from outside the game, or none", 0, 1, small, big)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "take nothing", map[string]any{}),
		choiceMove(t, 0, choiceID, "take Bear", map[string]any{"card_names": []string{"Bear"}}),
		choiceMove(t, 0, choiceID, "take Giant", map[string]any{"card_names": []string{"Giant"}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take Giant" {
		t.Fatalf("chose %q, want the best card from the collection", got)
	}
}
