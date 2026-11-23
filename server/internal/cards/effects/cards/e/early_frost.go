package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Early Frost — Instant {1}{U} (Fifth Dawn, #5604):
//
//	"Tap up to three target lands."
//
// Choking Tethers' tap of each still-legal target, over up to three lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f9cdbc2-d2f2-4390-8506-350fabf3f89a",
		Name:         "Early Frost",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("up to three target lands", effects.Land()).WithCount(0, 3),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapEachLegalTarget(ctx.Game, item)
		},
	})
}
