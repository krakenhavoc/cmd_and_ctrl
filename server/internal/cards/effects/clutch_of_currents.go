package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clutch of Currents — Sorcery {U}:
//
//	"Return target creature to its owner's hand.
//	 Awaken 3—{4}{U}"
//
// ADR 0135 §3 (#2411): the bounce, then the awaken land (CR 702.113a).
// Each target is checked on its own (CR 608.2b): a creature that left
// still lets the land awaken, and a land that left still lets the
// creature go home.
//
// No simplifications.
func init() {
	t := TargetCreature("target creature")
	Register(Spec{
		OracleID:     "2c6658c3-8c0e-46c9-8127-bc5b77eb0dab",
		Name:         "Clutch of Currents",
		Completeness: CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			Awaken(3, "{4}{U}", t),
		},
		OnResolve: AwakenAfter(3, func(_ *game.StackItem, ctx *Context) error {
			if id, ok := awakenSpellCardTarget(ctx); ok {
				return BounceToHand{Target: id}.Apply(ctx)
			}
			return nil
		}),
	})
}
