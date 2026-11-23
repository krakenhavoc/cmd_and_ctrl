package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vivisection — Sorcery {3}{U}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw three cards."
//
// Altar's Reap's machinery at three cards: the sacrifice is an
// additional cost paid on cast (CR 601.2b, 118.8), the draw is the
// whole effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "46f78c42-85d1-4cd4-9fc3-d5fdec04a4c8",
		Name:           "Vivisection",
		Purpose:        game.Purpose{Draws: 3, Answers: game.AnswerSacOutlet},
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
