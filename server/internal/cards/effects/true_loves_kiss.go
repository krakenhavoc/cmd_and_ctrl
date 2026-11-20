package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// True Love's Kiss — Instant {2}{W}{W} (Throne of Eldraine, #5516):
//
//	"Exile target artifact or enchantment.
//	 Draw a card."
//
// Altar's Light's exile, then the draw as a second step (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9506aba2-5e1d-48c5-ace8-4beb824ad8af",
		Name:         "True Love's Kiss",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return ExileFirstTarget(ctx.Game, ctx.Item) }),
			DrawCards{N: 1},
		),
	})
}
