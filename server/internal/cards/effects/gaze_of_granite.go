package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gaze of Granite — Sorcery {X}{B}{B}{G}:
//
//	"Destroy each nonland permanent with mana value X or less."
//
// X is the value chosen on cast, and the sweep reads each permanent's
// mana value on the battlefield, where an {X} in a cost counts as 0
// (CR 202.3e) and a token has mana value 0 — so X = 0 still destroys
// tokens and free permanents. The sweep goes through
// DestroyAllMatching, so the deaths are simultaneous and an
// indestructible permanent survives. It hits the caster's own
// permanents too, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "103d9ad0-d655-4bd5-a899-e9f8869e333d",
		Name:     "Gaze of Granite",
		// ADR 0126 §6: only mana value X or less.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Nonland(), ManaValueLE(ctx.X()))}.Apply(ctx)
		},
	})
}
