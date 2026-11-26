package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Into the Void — Sorcery {3}{U} (Conspiracy: Take the Crown, S82 older
// sets batch 271, #5540):
//
//	"Return up to two target creatures to their owners' hands."
//
// Aven Augur's bounce of every still-legal target, as a spell: none chosen
// is a legal announcement, and a target that is gone by resolution is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d5a970b-dd36-4871-a2aa-36ba21917cdf",
		Name:         "Into the Void",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove | game.AnswerSurvive},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceEachLegalTarget(ctx.Game, item)
		},
	})
}
