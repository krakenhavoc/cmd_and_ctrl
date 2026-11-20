package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grotesque Demise — Instant {2}{B} (Ravnica Allegiance, #5521):
//
//	"Exile target creature with power 3 or less."
//
// Exile on a power-gated target: the bound is a target predicate, judged
// at announce and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eed4651b-4c45-4136-9e3d-44471031368f",
		Name:         "Grotesque Demise",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with power 3 or less", PowerLE(3)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
