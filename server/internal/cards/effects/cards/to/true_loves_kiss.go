package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// True Love's Kiss — Instant {2}{W}{W} (Throne of Eldraine, #5516):
//
//	"Exile target artifact or enchantment.
//	 Draw a card."
//
// Altar's Light's exile, then the draw as a second step (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9506aba2-5e1d-48c5-ace8-4beb824ad8af",
		Name:         "True Love's Kiss",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.ExileFirstTarget(ctx.Game, ctx.Item) }),
			effects.DrawCards{N: 1},
		),
	})
}
