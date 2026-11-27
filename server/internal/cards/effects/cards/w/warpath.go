package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warpath — Instant {3}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Warpath deals 3 damage to each blocking creature and each blocked
//	 creature."
//
// Electryte's "each blocking creature", widened with the creatures that are
// blocked. Both sets are read as the spell resolves, a creature that is both
// blocked and blocking is dealt damage once, and all of it is one damage
// instance. It is castable any time, so with no combat it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "413520c7-5227-406f-8861-e844c153e5b0",
		Name:         "Warpath",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Or(effects.BlockingCreature(), effects.BlockedCreature()), 3)
		},
	})
}
