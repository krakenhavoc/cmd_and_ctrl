package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// --- #2523 / #2524: a choose_cards prompt over the bot's own graveyard

// graveyardChoice is a choose_cards prompt over seat 0's own graveyard —
// the shape Colossal Grave-Reaver and Eerie Ultimatum queue.
func graveyardChoice(id, reason string, lo, hi int, opts ...protocol.CardView) protocol.PendingChoiceView {
	return protocol.PendingChoiceView{
		ID: id, Kind: "choose_cards", Chooser: seatID(0).String(),
		FromPlayer: seatID(0).String(), Reason: reason,
		Options: opts, ChooseMin: lo, ChooseMax: hi, Count: hi,
	}
}

func TestGraveyardPickTakesTheBestCards(t *testing.T) {
	const choiceID = "choice-yard"
	dragon := creature(cardID(1), 0, "Dragon", 6, 6)
	bear := creature(cardID(2), 0, "Bear", 2, 2)

	// Colossal Grave-Reaver: exactly one, the better card wins whatever
	// order the enumerator lists them in.
	t.Run("exactly one: the best", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withGraveyard(bear, dragon)), newSeat(1)},
			withChoice(graveyardChoice(choiceID, "Colossal Grave-Reaver — put one of the milled creature cards onto the battlefield", 1, 1, bear, dragon)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "the bear", ids(cardID(2))),
			choiceMove(t, 0, choiceID, "the dragon", ids(cardID(1))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "the dragon" {
			t.Fatalf("chose %q, want the Dragon", got)
		}
	})

	// Eerie Ultimatum: "any number" with the empty answer offered first
	// must not win — reanimating everything legal does.
	t.Run("any number: takes the largest legal set, not nothing", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withGraveyard(bear, dragon)), newSeat(1)},
			withChoice(graveyardChoice(choiceID, "Eerie Ultimatum — return any number of permanent cards", 0, 2, bear, dragon)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "choose nothing", ids()),
			choiceMove(t, 0, choiceID, "the bear", ids(cardID(2))),
			choiceMove(t, 0, choiceID, "both", ids(cardID(1), cardID(2))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "both" {
			t.Fatalf("chose %q, want both cards", got)
		}
	})
}
