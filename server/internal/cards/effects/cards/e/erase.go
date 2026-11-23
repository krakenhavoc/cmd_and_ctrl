package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Erase — Instant {W} (Urza's Legacy, #5623):
//
//	"Exile target enchantment."
//
// Altar's Light narrowed to enchantments: the target is exiled, not
// destroyed, so an indestructible or regenerating enchantment goes too. An
// enchantment gone in response leaves the spell with nothing to do
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2ceb15c-5d02-4cf4-a7c9-c1a40b7ca667",
		Name:         "Erase",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
