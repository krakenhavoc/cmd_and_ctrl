package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hibernation — Instant {2}{U} (Eighth Edition, S82 older sets batch 618, #5607):
//
//	"Return all green permanents to their owners' hands."
//
// Crush of Tentacles' mass bounce with a colour test. Green is read through
// the layered view, so a permanent another effect made green is returned and
// one that lost its colour is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "633b541b-e37f-451a-be73-3398e61e0f75",
		Name:         "Hibernation",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.OfColor("G")}.Apply(ctx)
		},
	})
}
