package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inundate — Sorcery {3}{U}{U}{U} (Eventide, #5590):
//
//	"Return all nonblue creatures to their owners' hands."
//
// Evacuation over the nonblue creatures, their colour read as it resolves.
// A blue creature stays, so the sweep is partial (ADR 0126 §6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0390a3dc-386c-402f-a89c-b94867bfb0a2",
		Name:         "Inundate",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BounceAllMatching{Match: And(Creature(), Not(OfColor("U")))}.Apply(ctx)
		},
	})
}
