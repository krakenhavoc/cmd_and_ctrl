package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mending Touch — Instant {G}:
//
//	"Regenerate target creature."
//
// Refresh's regeneration without its draw: one regeneration shield on the
// target (CR 701.19a), used the next time it would be destroyed this turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3172ac0-213c-4da4-8b7a-8f84bc237cde",
		Name:         "Mending Touch",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return regenerateTheTargetPermanent(ctx.Game, item)
		},
	})
}
