package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// entry_reveal_test.go — #1198's bot branch, and the reason it is a
// branch of its own rather than a second spelling of choose_cards.
//
// The sign is the whole decision (#798). A card named to a
// choose_cards prompt over the bot's own hand is a card GIVEN UP —
// that branch scores an answer by what it KEEPS — and #1028's fuel
// pricer prices a card EATEN by a cost. A card named here is revealed
// and stays in hand (CR 701.20b): it costs nothing, and it buys an
// untapped land. Through either of the other two valuations the bot
// would decline every time and play every reveal-land tapped for the
// rest of the game, which is the regression this pins.

func TestEntryRevealPrefersRevealingOverEnteringTapped(t *testing.T) {
	const choiceID = "choice-reveal"
	swamp := land(cardID(1), 0)

	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withChoice(protocol.PendingChoiceView{
			ID:        choiceID,
			Kind:      "entry_reveal_from_hand",
			Chooser:   seatID(0).String(),
			Reason:    "Choked Estuary — reveal an Island or Swamp card from your hand?",
			Options:   []protocol.CardView{swamp},
			ChooseMin: 0,
			ChooseMax: 1,
		}))
	// The enumerator offers the decline FIRST, so a policy with no
	// opinion takes it. That is exactly the tie this branch exists to
	// break.
	in := input(0, v,
		choiceMove(t, 0, choiceID, "reveal nothing", map[string]any{"card_ids": []string{}}),
		choiceMove(t, 0, choiceID, "reveal the Swamp", map[string]any{"card_ids": []string{cardID(1)}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "reveal the Swamp" {
		t.Fatalf("chose %q, want the reveal — it costs nothing and the land enters untapped", got)
	}
}

// TestEntryRevealStillAnswersWithNothingToReveal: the decline is the
// only offer when the hand holds nothing the clause admits, and a
// seat owing a choice is offered nothing else — so the policy must
// still return an answer rather than stalling (#544).
func TestEntryRevealStillAnswersWithNothingToReveal(t *testing.T) {
	const choiceID = "choice-reveal-empty"
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withChoice(protocol.PendingChoiceView{
			ID:      choiceID,
			Kind:    "entry_reveal_from_hand",
			Chooser: seatID(0).String(),
			Reason:  "Port Town — reveal a Plains or Island card from your hand?",
		}))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "reveal nothing", map[string]any{"card_ids": []string{}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "reveal nothing" {
		t.Fatalf("chose %q, want the only offered answer", got)
	}
}
