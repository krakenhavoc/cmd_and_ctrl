package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Emissary — Creature — Human Wizard {2}{B}, 3/1 (Invasion, S82 OLD
// #5616):
//
//	"Kicker {1}{U} (You may pay an additional {1}{U} as you cast this spell.)
//	 When this creature enters, if it was kicked, return target permanent to
//	 its owner's hand."
//
// Mold Shambler's shape: an enters trigger that exists only when the spell
// was kicked (CR 702.33d, 603.4), its target chosen as it goes on the stack
// and judged again as it resolves (CR 608.2b). Any permanent is a legal
// target, the Emissary itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f30943f1-473f-4fa4-b6a7-3dfbd1995ba9",
		Name:          "Urborg Emissary",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Urborg Emissary — kicked, return target permanent to its owner's hand",
					effects.RetractionHelixBounceTarget),
				effects.TargetPermanent("target permanent")),
		},
	})
}
