package mcpseat

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0129 §7: a move that pays energy says how much, as one that pays
// life does, since the label's "Pay {E}{E}" is the only other place it
// shows.
func TestMoveLineSaysTheEnergyItCosts(t *testing.T) {
	v := &protocol.GameView{}
	m := legal.Move{Label: "Bristling Hydra: Pay {E}{E}{E}: …", Cost: &legal.MoveCost{Energy: 3}}
	got := moveLine(0, m, v, "", newNameWrapper(v))
	if !strings.Contains(got, "costs 3 energy") {
		t.Errorf("move line %q does not name the energy", got)
	}
}
