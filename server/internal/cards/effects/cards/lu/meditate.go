package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meditate — Instant {2}{U}:
//
//	"Draw four cards. You skip your next turn."
//
// Two sentences, in printed order: the draw, then the skip (CR 614.10,
// ADR 0188 Decision 1, #3565). The skip is a debt the rotation seam
// spends when the caster's next turn would begin, so casting it on your
// own turn never cuts that turn short, and casting it in response on an
// opponent's turn costs the caster the turn after it. Two Meditates are
// two skipped turns (CR 614.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f79f84c9-2023-4348-863b-a3b4bc382904",
		Name:         "Meditate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4, SkipsTurns: 1},
		// The skip waits for a prompt the draws ask (a dredge offer; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: 4}.Apply(ctx)
			}),
			effects.SkipYourNextTurns{},
		),
	})
}
