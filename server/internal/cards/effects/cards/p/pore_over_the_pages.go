package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pore Over the Pages — Sorcery {3}{U}{U} (Shadows over Innistrad, #5543):
//
//	"Draw three cards, untap up to two lands, then discard a card."
//
// Frantic Search's steps in the printed order: the draw, then the untap
// (a prompt over every tapped land at the table, since the text names no
// controller), then the discard the controller chooses. Each step waits for
// the prompt of the one before it (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55fe5ec3-1c81-4e9d-9e4c-f68ce6d7c2de",
		Name:         "Pore Over the Pages",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 1, Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 3},
			effects.UntapUpToLands{N: 2, Question: "Pore Over the Pages — untap up to two lands"},
			effects.Instruction(func(ctx *effects.Context) error {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: ctx.Controller(), Source: ctx.Source(), N: 1})
				return nil
			}),
		),
	})
}
