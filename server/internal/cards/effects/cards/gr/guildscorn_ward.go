package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guildscorn Ward — Enchantment — Aura {W} (Gatecrash, S82 OLD b407, #5565):
//
//	"Enchant creature
//	 Enchanted creature has protection from multicolored."
//
// Spirit Mantle's keyword grant with Argentum Masticore's "protection from
// multicolored" (damage, enchanting, blocking and targeting by multicolored
// sources, CR 702.16b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "daff56a6-a5b6-4726-8fee-676b846e0c6b",
		Name:         "Guildscorn Ward",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("protection from multicolored")},
	})
}
