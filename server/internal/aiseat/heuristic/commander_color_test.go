package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// commander_color_test.go — CR 903.4b (#5948): a deck that shows no
// colour for its Prismatic Piper is asked, as the mulligan window closes,
// with the ordinary choose_color prompt, declaring no purpose. The board is
// empty then, so the bot names the colour its kept hand asks for.
func TestABotNamesItsCommandersColourFromItsOpeningHand(t *testing.T) {
	hand := withHand(
		colored(spell(cardID(90), 0, "Lightning Bolt", "{R}"), "R"),
		colored(spell(cardID(91), 0, "Fireball", "{X}{R}"), "R"),
		colored(spell(cardID(92), 0, "Giant Growth", "{G}"), "G"),
	)
	v := newView([]protocol.PlayerView{newSeat(0, hand), newSeat(1)},
		withChoice(protocol.PendingChoiceView{
			ID:           colorChoiceID,
			Kind:         "choose_color",
			Chooser:      seatID(0).String(),
			FromPlayer:   seatID(0).String(),
			Count:        1,
			Reason:       "The Prismatic Piper is your commander: choose its color",
			ColorOptions: []string{"W", "U", "B", "R", "G"},
		}))
	in := input(0, v, colorMoves(t, 0)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose R" {
		t.Errorf("chose %q for its commander, want the colour its hand asks for most (R)", got)
	}
}
