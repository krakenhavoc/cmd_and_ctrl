package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volatile Claws — Instant {2}{R} (Modern Horizons):
//
//	"Until end of turn, creatures you control get +2/+0 and gain all creature
//	 types."
//
// The creatures it affects are fixed as the spell resolves (CR 611.2c): a
// creature that enters later this turn is neither pumped nor a changeling.
// "Gain all creature types" is the changeling mod (CR 702.73a, CR 205.3m),
// layer 4, so the creatures count as every type for the turn; the pump is
// layer 7c.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a064c10-5bd1-41ec-b2b5-d0eb33c73c9b",
		Name:         "Volatile Claws",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.YouControl()),
				"Volatile Claws — +2/+0 and all creature types",
				game.ModifyPTMod(2, 0), game.AllCreatureTypesMod())
		},
	})
}
