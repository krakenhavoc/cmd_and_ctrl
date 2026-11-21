package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// leaveDeclareAttackers ends the active player's declaration through
// AdvanceStep, which runs CR 508.1d's checkpoint.
func LeaveDeclareAttackers(g *game.Game) error {
	_, err := g.AdvanceStep()
	return err
}
