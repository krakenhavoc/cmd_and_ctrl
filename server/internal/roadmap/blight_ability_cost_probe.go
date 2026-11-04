package roadmap

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// declaresAbilityBlight is the "Blight N" ability-cost probe (#4684): an
// activated ability whose cost blights (game.AbilityCost.Blight, built by
// effects.BlightN).
func declaresAbilityBlight(s effects.Spec) bool {
	for _, a := range s.Activated {
		if a.Cost.Blight > 0 {
			return true
		}
	}
	return false
}
