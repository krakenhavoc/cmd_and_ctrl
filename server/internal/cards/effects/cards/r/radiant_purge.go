package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Radiant Purge — Instant {1}{W} (Dragons of Tarkir):
//
//	"Exile target multicolored creature or multicolored enchantment."
//
// Multicolored is read from the permanent's current colours (CR 105.4, two or
// more).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b61718ad-7102-4ed2-b884-733bc46df715",
		Name:         "Radiant Purge",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target multicolored creature or multicolored enchantment",
			effects.And(effects.Multicolored(), effects.Or(effects.Creature(), effects.Enchantment()))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
