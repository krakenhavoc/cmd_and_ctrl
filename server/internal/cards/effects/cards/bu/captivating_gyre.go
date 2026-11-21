package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captivating Gyre — Sorcery {4}{U}{U} (Core Set 2020, #5518):
//
//	"Return up to three target creatures to their owners' hands."
//
// Each pick still legal as it resolves goes to its owner's hand
// (bounceEachLegalTarget); choosing none is legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41011f09-ad8a-4eea-aa52-d20a219bc94c",
		Name:         "Captivating Gyre",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceEachLegalTarget(ctx.Game, item)
		},
	})
}
