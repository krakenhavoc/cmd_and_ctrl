package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glare of Heresy — Sorcery {1}{W} (Theros, #5561):
//
//	"Exile target white permanent."
//
// Celestial Purge's body for a white permanent of any type.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e127e6e-b292-4398-b3fa-8b520a7d26a3",
		Name:         "Glare of Heresy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target white permanent", effects.OfColor("W")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
