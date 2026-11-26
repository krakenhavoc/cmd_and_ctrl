package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB347XTargetCreatures is "X target creatures": a creature clause whose
// count is the announced X (Icy Blast).
func OldB347XTargetCreatures() *game.TargetSpec {
	spec := TargetCreature("X target creatures")
	spec.CountFromX = true
	return spec
}
