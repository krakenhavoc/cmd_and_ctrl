package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "05a1575c-8d08-4845-8ae0-85e0aba55637",
		Name:         "Clear a Path",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with defender", effects.HasKeyword("defender")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}
