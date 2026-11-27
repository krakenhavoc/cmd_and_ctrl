package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dying Wail — {1}{B} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 When enchanted creature dies, target player discards two cards."
//
// The target player chooses what to discard (Mournwhelk's discard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78a9cb8d-52bb-4182-ad7d-e4f8f0bb9797",
		Name:         "Dying Wail",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenEnchantedCreatureDies("Dying Wail — target player discards two cards", effects.TargetPlayerDiscards(2)),
				effects.TargetPlayer("target player")),
		},
	})
}
