package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soothing Balm — Instant {1}{W} (Mercadian Masques, #5619):
//
//	Target player gains 5 life.
//
// The target is re-checked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f493e26-a556-4ca0-bb44-5408800ccf82",
		Name:         "Soothing Balm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			who := effects.FirstLegalTargetPlayer(ctx)
			if who == uuid.Nil {
				return nil
			}
			return effects.GainLife{Player: who, Amount: 5}.Apply(ctx)
		},
	})
}
