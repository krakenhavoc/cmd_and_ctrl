package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raphael's Technique — Instant {4}{R}{R} (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Sneak {2}{R} (You may cast this spell for {2}{R} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step.)
//	 Each player may discard their hand and draw seven cards."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The wheel is
// Ruin Grinder's: each player, starting with the caster, chooses for
// themselves (eachPlayerMayWheel).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f58a5f8b-254c-4206-bdbe-55a23798ed2d",
		Name:             "Raphael's Technique",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{2}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return eachPlayerMayWheel(ctx, "Raphael's Technique — discard your hand and draw seven cards?")
		},
	})
}
