package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Organic Extinction — Sorcery {8}{W}{W}: "Improvise ... Destroy all
// nonartifact creatures."
//
// Improvise is Spec.TapCost; the sweep is DestroyAllMatching so the
// deaths are simultaneous. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9e24d7c-0907-45c1-b5d3-9fb16a929868",
		Name:         "Organic Extinction",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Improvise(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.Not(effects.Artifact()))}.Apply(ctx)
		},
	})
}
