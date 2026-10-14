package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0189 decision 3 (#3538): a choose_cards prompt tagged "manifest"
// is Kozilek, the Broken Reality (or Scroll of Fate) asking the bot
// which of its own hand cards to put onto the battlefield face down. A
// face-down 2/2 is not the card, so the bot gives up its worst cards:
// the answer is priced by what it keeps, like a discard, and not as a
// put (puts.go). The winning answer is listed last, so the enumerator's
// order decides nothing.
func TestTheBotManifestsItsWorstCards(t *testing.T) {
	const choiceID = "choice-manifest"
	dragon := creature(cardID(1), 0, "Dragon", 6, 6)
	bear := creature(cardID(2), 0, "Bear", 2, 2)
	mountain := land(cardID(3), 0)

	pick := handChoice(choiceID, "Choose 2 cards from your hand to manifest", 2, 2, dragon, bear, mountain)
	pick.ChooseDestination = "manifest"
	v := newView([]protocol.PlayerView{newSeat(0, withHand(dragon, bear, mountain)), newSeat(1)},
		withBattlefield(sixLands()...), withChoice(pick))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "manifest dragon + bear", map[string]any{"card_ids": []string{cardID(1), cardID(2)}}),
		choiceMove(t, 0, choiceID, "manifest dragon + mountain", map[string]any{"card_ids": []string{cardID(1), cardID(3)}}),
		choiceMove(t, 0, choiceID, "manifest bear + mountain", map[string]any{"card_ids": []string{cardID(2), cardID(3)}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "manifest bear + mountain" {
		t.Fatalf("chose %q, want the answer that keeps the Dragon in hand", got)
	}
}
