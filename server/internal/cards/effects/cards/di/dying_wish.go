package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dying Wish — Enchantment — Aura {1}{B} (Gatecrash, S82 OLD b407, #5565):
//
//	"Enchant creature you control
//	 When enchanted creature dies, target player loses X life and you gain X
//	 life, where X is its power."
//
// Murder Investigation's dies trigger with a player target chosen as the
// trigger goes on the stack (CR 603.3d). X is the creature's last-known power
// (CR 603.10a). The life loss and the gain are one instruction apiece and only
// happen if the target player is still legal as the trigger resolves
// (CR 608.2b); X of zero or less does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "302407c8-0842-4f8e-a5aa-5c11a69851c4",
		Name:         "Dying Wish",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenEnchantedCreatureDies("Dying Wish — target player loses X life and you gain X life, where X is its power",
				effects.S82OldB407DyingWishDrain), effects.TargetPlayer("target player")),
		},
	})
}
