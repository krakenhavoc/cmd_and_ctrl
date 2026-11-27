package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aether Sting — {3}{R} Enchantment (Urza's Destiny, S82 OLD batch 704, #5621):
//
//	"Whenever an opponent casts a creature spell, this enchantment deals 1
//	 damage to that player."
//
// One trigger per creature spell an opponent casts (Weeping Angel's cast
// condition); the damage is the enchantment's, to the caster.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61105cb5-d7a1-4021-a006-dd1b947dfa68",
		Name:         "Aether Sting",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.Creature()),
				"Aether Sting — it deals 1 damage to that player", effects.S82OldB704AetherStingDamage),
		},
	})
}
