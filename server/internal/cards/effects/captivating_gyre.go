package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captivating Gyre — Sorcery {4}{U}{U} (Core Set 2020, #5518):
//
//	"Return up to three target creatures to their owners' hands."
//
// Each pick still legal as it resolves goes to its owner's hand
// (bounceEachLegalTarget); choosing none is legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41011f09-ad8a-4eea-aa52-d20a219bc94c",
		Name:         "Captivating Gyre",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return bounceEachLegalTarget(ctx.Game, item)
		},
	})
}
