package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planar Outburst — Sorcery {3}{W}{W}:
//
//	"Destroy all nonland creatures.
//	 Awaken 4—{5}{W}{W}{W}"
//
// ADR 0135 §3 (#2411): the sweep spares land creatures, an awakened land
// among them; then the awaken land (CR 702.113a). Regeneration applies
// (the card does not say it can't).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ff0d05b9-0c9c-4217-abd2-7e7757a0c4c9",
		Name:         "Planar Outburst",
		Completeness: CompletenessFull,
		// Partial: a land creature is spared.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		AlternativeCosts: []game.AlternativeCost{
			Awaken(4, "{5}{W}{W}{W}", nil),
		},
		OnResolve: AwakenAfter(4, func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), Not(Land()))}.Apply(ctx)
		}),
	})
}
