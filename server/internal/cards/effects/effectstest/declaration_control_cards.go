package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func DeclarationRecords(g *game.Game) []game.DeclarationControl {
	var out []game.DeclarationControl
	g.ReadSnapshot(func() { out = g.DeclarationControlsForEffect() })
	return out
}
