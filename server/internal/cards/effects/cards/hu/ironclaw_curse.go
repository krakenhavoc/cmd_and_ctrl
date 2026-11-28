package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ironclaw Curse — Enchantment — Aura {R} (Homelands, S82 OLD b825, #5635):
//
//	"Enchant creature
//	 Enchanted creature gets -0/-1.
//	 Enchanted creature can't block creatures with power equal to or greater
//	 than the enchanted creature's toughness."
//
// The -0/-1 is a layer 7c modify; the block rule is a blocker-side pair rule
// that reads both creatures live, through the layers, as the block is
// declared (CR 509.1b), so the Curse's own -0/-1 counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51708bd1-23db-4597-84a7-f36abd93c537",
		Name:         "Ironclaw Curse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(0, -1)},
		BlockRules: []game.BlockRule{
			effects.EnchantedCreatureCantBlockPowerAtLeastItsToughness("creatures with power equal to or greater than the enchanted creature's toughness"),
		},
	})
}
