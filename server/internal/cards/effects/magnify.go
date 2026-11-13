package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magnify — Instant {G} (Urza's Destiny, #5621):
//
//	"All creatures get +1/+1 until end of turn."
//
// Marsh Gas's shape the other way: every creature on the battlefield as
// the spell resolves, every player's, gets +1/+1 until the cleanup
// step. A creature that arrives later in the turn is not affected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "421c661a-24d9-4797-b538-2a4243825628",
		Name:         "Magnify",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{Match: Creature(), Power: 1, Toughness: 1, Label: "Magnify — all creatures get +1/+1"}.Apply(ctx)
		},
	})
}
