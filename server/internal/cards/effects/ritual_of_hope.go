package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ritual of Hope — Instant {1}{W} (Midnight Hunt, #5497):
//
//	"Creatures you control get +1/+1 until end of turn.
//	 Coven — If you control three or more creatures with different
//	 powers, creatures you control get +2/+1 until end of turn instead."
//
// Coven is read once as the spell resolves, and the creatures it affects
// are fixed then (CR 611.2c); it is one boost or the other, never both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8afe3d28-be3e-416c-9457-8a2af85b2eec",
		Name:         "Ritual of Hope",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			power := 1
			if trekB48Coven(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID) {
				power = 2
			}
			return BoostUntilEOT{
				Match:     And(Creature(), YouControl()),
				Power:     power,
				Toughness: 1,
				Label:     "Ritual of Hope — creatures you control get " + signedPT(power, 1) + " until end of turn",
			}.Apply(ctx)
		})),
	})
}
