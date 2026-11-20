package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sinister Sabotage — Instant {1}{U}{U} (Guilds of Ravnica, #5525):
//
//	"Counter target spell.
//	 Surveil 1."
//
// Two steps: Cancel's counter (counterTheTargetSpell), then surveil 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "973fd4d4-9255-4825-85b7-503606c4e932",
		Name:         "Sinister Sabotage",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return counterTheTargetSpell(ctx.Item, ctx) }),
			Surveil{N: 1},
		),
	})
}
