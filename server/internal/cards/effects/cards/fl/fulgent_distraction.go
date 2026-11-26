package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fulgent Distraction — Instant {2}{W} (Scars of Mirrodin, S82 OLD batch 471,
// #5577):
//
//	"Choose two target creatures. Tap those creatures, then unattach all
//	 Equipment from them."
//
// Exactly two creature targets. Each one still legal as it resolves is tapped
// (CR 608.2b), then every Equipment attached to one of them is unattached
// (CR 701.3d) and stays on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7996904a-1d7c-4d95-8fd8-431272c7c7e0",
		Name:         "Fulgent Distraction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.TapEachLegalTarget(ctx.Game, item); err != nil {
				return err
			}
			return effects.OldB471UnattachEquipmentFrom(ctx, effects.HoldTargetIDs(ctx))
		},
	})
}
