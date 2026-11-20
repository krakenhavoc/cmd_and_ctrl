package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clear a Path — Sorcery {R}:
//
//	"Destroy target creature with defender."
//
// Smash to Dust's second bullet as a whole spell: the defender clause is a
// target predicate, judged at announce (CR 601.2c) and again at resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05a1575c-8d08-4845-8ae0-85e0aba55637",
		Name:         "Clear a Path",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with defender", HasKeyword("defender")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyChosenPermanent(ctx.Game, item)
		},
	})
}
