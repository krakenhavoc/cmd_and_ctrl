package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// neoB01ControlsAnEnchantment is "you control an enchantment", read after
// the layers so a permanent made an enchantment by another effect counts
// (Cloudsculpt Technician's check for artifacts).
func NeoB01ControlsAnEnchantment(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsEnchantment() {
			return true
		}
	}
	return false
}
