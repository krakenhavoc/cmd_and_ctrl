package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benalish Emissary — Creature — Human Wizard {2}{W}, 1/4 (Invasion, S82 OLD
// #5616):
//
//	"Kicker {1}{G} (You may pay an additional {1}{G} as you cast this spell.)
//	 When this creature enters, if it was kicked, destroy target land."
//
// Mold Shambler's shape: an enters trigger that exists only when the spell
// was kicked (CR 702.33d, 603.4), its target chosen as it goes on the stack
// and judged again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "99f2e195-5da5-4239-87e0-ed552643e7b5",
		Name:          "Benalish Emissary",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Benalish Emissary — kicked, destroy target land",
					effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target land", effects.Land())),
		},
	})
}
