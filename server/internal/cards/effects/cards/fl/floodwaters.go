package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Floodwaters — {4}{U}{U} Sorcery (Amonkhet, S82 old batch 245):
//
//	"Return up to two target creatures to their owners' hands.
//	 Cycling {2}"
//
// Each target still legal as it resolves returns to its owner's hand
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e89cce0e-4fd7-40fa-8a55-b07d72826956",
		Name:         "Floodwaters",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove | game.AnswerSurvive},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceEachLegalTarget(ctx.Game, item)
		},
	})
}
