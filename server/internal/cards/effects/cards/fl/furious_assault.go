package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Furious Assault — Enchantment {2}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Whenever you cast a creature spell, this enchantment deals 1 damage to
//	 target player or planeswalker."
//
// The target is chosen as the trigger goes on the stack, and the damage is the
// enchantment's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5d1a004-5598-4201-ac07-bca6cd23008f",
		Name:         "Furious Assault",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.Creature(), "Furious Assault — 1 damage to target player or planeswalker",
				effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetPlayerOrPlaneswalker()),
		},
	})
}
