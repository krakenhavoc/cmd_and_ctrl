package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scent of Cinder — {1}{R} Sorcery (Urza's Destiny, S82 OLD batch 704, #5621):
//
//	"Reveal any number of red cards in your hand. Scent of Cinder deals X
//	 damage to any target, where X is the number of cards revealed this way."
//
// Cinder Seer's reveal-and-damage as a spell (S82OldB704RevealRedForDamage).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff50d002-834e-4b5e-8db0-b0c8b4efb21f",
		Name:         "Scent of Cinder",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB704RevealRedForDamage("Scent of Cinder")(ctx.Game, item)
		},
	})
}
