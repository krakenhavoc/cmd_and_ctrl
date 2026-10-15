package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Organic Extinction — Sorcery {8}{W}{W}: "Improvise ... Destroy all
// nonartifact creatures."
//
// Improvise is Spec.TapCost; the sweep is DestroyAllMatching so the
// deaths are simultaneous. No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9e24d7c-0907-45c1-b5d3-9fb16a929868",
		Name:         "Organic Extinction",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		TapCost:      Improvise(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), Not(Artifact()))}.Apply(ctx)
		},
	})
}
