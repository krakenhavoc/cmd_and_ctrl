package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strength of Unity — Enchantment — Aura {3}{W} (Invasion, #5616):
//
//	"Enchant creature
//	 Domain — Enchanted creature gets +1/+1 for each basic land type among
//	 lands you control."
//
// A layer 7c static on the host (CR 613.4c) whose count is the Aura's
// controller's domain, read on every pass, so it follows the lands.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB660DomainAura("00606d5e-a134-4363-9ee6-05e03d58f234", "Strength of Unity", 1, game.AuraBenefit))
}
