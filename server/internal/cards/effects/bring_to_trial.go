package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bring to Trial — Sorcery {2}{W} (Ravnica Allegiance, #5521):
//
//	"Exile target creature with power 4 or greater."
//
// Exile on a power-gated target: the bound is a target predicate, judged
// at announce and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "968b6277-584c-4947-98ac-4d6e5f8d6754",
		Name:         "Bring to Trial",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with power 4 or greater", PowerGE(4)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
