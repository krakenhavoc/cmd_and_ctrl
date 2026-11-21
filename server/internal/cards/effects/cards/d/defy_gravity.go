package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Defy Gravity — Instant {U} (Judgment, #5611):
//
//	"Target creature gains flying until end of turn.
//	 Flashback {U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Leap's flying without the draw: a layer-6 grant that ends at cleanup
// (CR 514.2), declared as a benefit. A spell whose only target has gone
// is countered by the game rules (CR 608.2b). Flashback is the shared
// constructor: the graveyard cast, the price, and CR 702.34a's exile as
// the spell leaves the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "5a295dc1-7f60-4581-b0f9-61e8b14339bf",
		Name:             "Defy Gravity",
		Completeness:     effects.CompletenessFull,
		Purpose:          effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{U}")},
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Defy Gravity — flying until end of turn"}.Apply(ctx)
			}),
		),
	})
}
