package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kolaghan, the Storm's Fury — Legendary Creature — Dragon {3}{B}{R}, 4/5
// (Commander 2017, #5533):
//
//	"Flying
//	 Whenever a Dragon you control attacks, creatures you control get +1/+0
//	 until end of turn.
//	 Dash {3}{B}{R} (You may cast this spell for its dash cost. If you do, it
//	 gains haste, and it's returned from the battlefield to its owner's hand
//	 at the beginning of the next end step.)"
//
// Silumgar's attack trigger, one per attacking Dragon, over your own
// creatures as the ability resolves (CR 611.2c). Dash is the shared
// alternative cost (Kolaghan Skirmisher, CR 702.109): a dashed Kolaghan has
// haste and so can attack and trigger herself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "f6e160c1-b29f-4f3d-ab55-43ffe3ab4acf",
		Name:             "Kolaghan, the Storm's Fury",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{3}{B}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.DragonAttackTrigger("Kolaghan — creatures you control get +1/+0 until end of turn",
				func(ctx *effects.Context, _ uuid.UUID) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1,
						Label: "Kolaghan — +1/+0",
					}.Apply(ctx)
				}),
		},
	})
}
