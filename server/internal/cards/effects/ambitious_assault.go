package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambitious Assault — Instant {2}{R} (Kamigawa: Neon Dynasty):
//
//	"Creatures you control get +2/+0 until end of turn. If you control a
//	 modified creature, draw a card. (Equipment, Auras you control, and
//	 counters are modifications.)"
//
// The pump is applied first; the draw reads the board after it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3fb399fa-123f-448e-a473-573a1fab2c78",
		Name:         "Ambitious Assault",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Label: "Ambitious Assault"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !neoB10ControlsModified(ctx.Game, ctx.Controller()) {
					return nil
				}
				return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
