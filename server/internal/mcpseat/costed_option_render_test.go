package mcpseat

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #2854: an answer to an option_pick whose option costs mana says what
// it pays, since the server taps for it when the answer arrives.
func TestMoveLineSaysTheManaAnOptionPays(t *testing.T) {
	v := &protocol.GameView{}
	m := legal.Move{Kind: legal.KindChoice, Label: "Winter's Chill: Pay {2}: Bear is unaffected", Cost: &legal.MoveCost{Mana: "{2}"}}
	got := moveLine(0, m, v, "", newNameWrapper(v))
	if !strings.Contains(got, "pays {2}") {
		t.Errorf("move line %q does not name the payment", got)
	}
}
