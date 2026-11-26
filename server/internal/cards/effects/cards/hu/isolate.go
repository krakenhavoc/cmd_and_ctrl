package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Isolate — Instant {W} (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"Exile target permanent with mana value 1."
//
// Banish from Edoras's exile of the first legal target, over any permanent
// whose mana value is exactly 1 (a token or a land has mana value 0 and is not
// a target).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f42d4556-5be2-4861-9f88-1b630b43eff2",
		Name:         "Isolate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target permanent with mana value 1", effects.ManaValueGE(1), effects.ManaValueLE(1)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
