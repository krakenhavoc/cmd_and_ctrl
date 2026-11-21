package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alchemist's Retrieval — Instant {U}:
//
//	"Cleave {1}{U} (You may cast this spell for its cleave cost. If you
//	 do, remove the words in square brackets.)
//	 Return target nonland permanent [you control] to its owner's hand."
//
// Hard-cast it is a bounce for your own permanent; cleaved the target
// clause widens to any nonland permanent (CR 702.148a), which Cleave's
// second argument swaps in, so the announce check and the CR 608.2b
// re-check both read the clause that was paid for.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8651da8e-9b2c-4530-a521-aea490ad88d4",
		Name:         "Alchemist's Retrieval",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent you control", effects.Nonland(), effects.YouControl()),
		AlternativeCosts: []game.AlternativeCost{
			effects.Cleave("{1}{U}", effects.TargetPermanent("target nonland permanent", effects.Nonland())),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstTargetCard(item)
			if !ok {
				return nil
			}
			return effects.BounceToHand{Target: t.ID}.Apply(ctx)
		},
	})
}
