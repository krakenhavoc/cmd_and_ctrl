package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3fb399fa-123f-448e-a473-573a1fab2c78",
		Name:         "Ambitious Assault",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Label: "Ambitious Assault"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !effects.NeoB10ControlsModified(ctx.Game, ctx.Controller()) {
					return nil
				}
				return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
