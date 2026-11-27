package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Warcraft — Instant {2}{W} (Judgment, #5611):
//
//	"Creatures you control get +0/+2 until end of turn.
//	 Threshold — If there are seven or more cards in your graveyard, choose
//	 a color. Creatures you control also gain protection from the chosen
//	 color until end of turn."
//
// The pump and the colour grant are two steps (the second may stop to ask for
// the colour), over the creatures you control as the spell resolves; the
// threshold is read as it resolves, with the spell itself still on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c152f829-0452-407b-8020-023bf3d75c74",
		Name:         "Aven Warcraft",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CreaturesYouControlGetUntilEOT(0, 2, "Aven Warcraft — +0/+2")(ctx.Item, ctx)
			}),
			effects.OldB622AvenWarcraftThreshold),
	})
}
