package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liberate — Instant {1}{W} (Invasion, #5616):
//
//	"Exile target creature you control. Return that card to the battlefield
//	 under its owner's control at the beginning of the next end step."
//
// Baboon Spirit's delayed blink: the exile happens now and the return is a
// delayed trigger (CR 603.7) at the next end step; the card comes back as a
// new object (CR 400.7) and a token exiled this way does not return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f17e6c85-4813-4b0f-a0e9-d295822e9bb5",
		Name:         "Liberate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTargetsThenScheduleReturn(ctx, "Liberate — return the exiled creature")
		},
	})
}
