package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radstorm — Instant {3}{U} (Fallout, #5469):
//
//	"Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn.)
//	 Proliferate. (Choose any number of permanents and/or players, then
//	 give each another counter of each kind already there.)"
//
// Brain Freeze's storm row (Storm, CR 702.40) on a spell whose whole
// effect is the player's proliferate choice (Proliferate, #2525). Each
// copy resolves on its own, so each proliferate is chosen separately,
// after the counters the one before it added.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4cde3d5-eb13-4ad5-8178-08780dc85d39",
		Name:         "Radstorm",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Proliferate{}.Apply(ctx)
		},
	})
}
