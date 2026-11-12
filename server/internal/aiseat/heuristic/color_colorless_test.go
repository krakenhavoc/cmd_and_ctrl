package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// color_colorless_test.go — #5650: Angelic Intervention's prompt offers
// colorless beside the five colours, and the bot names it when the
// biggest thing pointed its way has no colour.

func TestProtectionNamesColorlessAgainstAColourlessAttacker(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			creature(cardID(3), 0, "Bear", 2, 2),
			creature(cardID(4), 1, "Their Golem", 6, 6, attacking(0)),
			colored(creature(cardID(5), 1, "Their Goblin", 2, 2, attacking(0)), "R")),
		withTurn(4, 1, "declare_attackers"), withColorChoice(0, "protect"))
	moves := append(colorMoves(t, 0), choiceMove(t, 0, colorChoiceID, "choose C", map[string]any{"color": "C"}))
	in := input(0, v, moves...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose C" {
		t.Errorf("the bot named %q, want colorless: the 6/6 attacker has no colour", got)
	}
}
