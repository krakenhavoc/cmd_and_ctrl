package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82B30ControlsArtifactOrEnchantment is "you control an artifact or
// enchantment" (Flutterfox), read after the layers.
func S82B30ControlsArtifactOrEnchantment(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && (c.IsArtifact() || c.IsEnchantment()) {
			return true
		}
	}
	return false
}
