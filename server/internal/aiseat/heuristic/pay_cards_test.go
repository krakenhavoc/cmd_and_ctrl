package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// pay_cards_test.go — ADR 0108 §5 decision 6: the bot pays an echo or
// cumulative upkeep whose cost is a sacrifice when it can, and gives up
// the least valuable permanents to do it.

func TestPayUnlessSacrificePaysWithTheLeastValuable(t *testing.T) {
	const choiceID = "choice-echo"
	island := land(cardID(1), 0)
	dragon := creature(cardID(2), 0, "Dragon", 6, 6)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(island, dragon, creature(cardID(3), 0, "Surger", 6, 4)),
		withChoice(protocol.PendingChoiceView{
			ID:       choiceID,
			Kind:     "pay_unless",
			Chooser:  seatID(0).String(),
			Reason:   "Phyrexian Soulgorger — cumulative upkeep: sacrifice a creature",
			PayCost:  "Sacrifice a creature",
			PayCards: &protocol.PayCardsView{Action: "sacrifice", Count: 1, Options: []string{cardID(1), cardID(2)}},
		}))
	yes, no := true, false
	in := input(0, v,
		choiceMove(t, 0, choiceID, "decline", map[string]any{"apply": no}),
		choiceMove(t, 0, choiceID, "pay with the dragon", map[string]any{"apply": yes, "card_ids": []string{cardID(2)}}),
		choiceMove(t, 0, choiceID, "pay with the island", map[string]any{"apply": yes, "card_ids": []string{cardID(1)}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pay with the island" {
		t.Fatalf("chose %q, want to pay with the least valuable permanent", got)
	}
}
