package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valgavoth's Onslaught — Sorcery {X}{X}{G}:
//
//	"Manifest dread X times, then put X +1/+1 counters on each of those
//	 creatures."
//
// X manifest dreads in a row, each its own look at the top two cards of
// the library as the earlier ones left it. "Each of those creatures" is
// the permanents that entered, so a manifest the library was too small
// to make, or one a replacement sent elsewhere, is not counted; a
// creature that left the battlefield before the counters go on is
// skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "663d9559-6619-46fe-8ff4-655f1d5e0e4c",
		Name:         "Valgavoth's Onslaught",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			return ManifestDreadTimes{N: x, Then: putCountersOnEachManifested(game.CounterPlusOne, x)}.Apply(ctx)
		},
	})
}
