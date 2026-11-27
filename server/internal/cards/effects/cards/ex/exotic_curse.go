package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exotic Curse — Enchantment — Aura {2}{B} (Invasion, #5616):
//
//	"Enchant creature
//	 Domain — Enchanted creature gets -1/-1 for each basic land type among
//	 lands you control."
//
// Strength of Unity's static with the sign turned round: the Aura's
// controller's domain, read on every pass.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB660DomainAura("344476b1-8710-4c7c-826d-efb7d13ceaa7", "Exotic Curse", -1, game.AuraHarm))
}
