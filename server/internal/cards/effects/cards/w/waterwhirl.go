package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waterwhirl — Instant {4}{U}{U} (Khans of Tarkir, #5554):
//
//	"Return up to two target creatures to their owners' hands."
//
// Whiplash Trap's simultaneous bounce over up to two targets: the ones
// still legal return together, and a target that left is skipped (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77636154-962f-4ca8-85e7-c0c9c0c3cd2d",
		Name:         "Waterwhirl",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ctx.Game.BounceCardsToHandForEffect(effects.LegalTargetCardIDs(ctx))
			return nil
		},
	})
}
