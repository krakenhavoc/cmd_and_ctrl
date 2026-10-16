package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dream's Grip — Instant {U}:
//
//	"Choose one —
//	 • Tap target permanent.
//	 • Untap target permanent.
//	 Entwine {1} (Choose both if you pay the entwine cost.)"
//
// Entwine (CR 702.42, #3866): {1} more takes both. Each bullet has its
// own "target", so entwined the same permanent may be chosen for both
// (CR 115.3): it is tapped, then untapped, in the order written (CR
// 702.42b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8b239262-c392-4ec6-9ae1-8b4a076e4bf0",
		Name:          "Dream's Grip",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{1}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Tap target permanent.", TargetPermanent("target permanent"), TapTheModesTarget),
			ModeDoing("Untap target permanent.", TargetPermanent("target permanent"),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return UntapTarget{Target: id}.Apply(ctx)
				})),
		)),
	})
}
