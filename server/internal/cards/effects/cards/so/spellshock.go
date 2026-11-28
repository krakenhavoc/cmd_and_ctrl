package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spellshock — Enchantment {2}{R} (Exodus, S82 OLD b738, #5626):
//
//	"Whenever a player casts a spell, this enchantment deals 2 damage to that
//	 player."
//
// Eidolon of the Great Revel without the mana-value gate: every spell cast by
// any player, the Enchantment's controller's included, deals 2 damage to its
// caster.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6579bb00-3886-447b-a265-3e72e98bbe2c",
		Name:         "Spellshock",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnyPlayer, "Spellshock — 2 damage to that player", effects.ScgB1DamageThatPlayer(2)),
		},
	})
}
