package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Opal Titan — Enchantment {2}{W}{W} (Urza's Saga, S82 OLD b721):
//
//	"When an opponent casts a creature spell, if this permanent is an
//	 enchantment, it becomes a 4/4 Giant creature with protection from each of
//	 that spell's colors."
//
// Same shape as the other hidden enchantments; the protection is one quality
// per color of the triggering spell, read as the trigger resolves (CR 702.16),
// and a colorless spell gives none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce06fe83-a8c5-46b7-b30c-b8189d217741",
		Name:         "Opal Titan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB721OpponentCastWhileEnchantment(effects.Creature()),
				"Opal Titan — it becomes a 4/4 Giant creature with protection from each of that spell's colors",
				effects.S82OldB721OpalTitanBecomes),
		},
	})
}
