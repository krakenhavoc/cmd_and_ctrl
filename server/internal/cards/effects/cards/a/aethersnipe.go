package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aethersnipe — Creature — Elemental {5}{U}, 4/4 (Lorwyn, #5593):
//
//	"When this creature enters, return target nonland permanent to its
//	 owner's hand.
//	 Evoke {1}{U}{U} (You may cast this spell for its evoke cost. If you do,
//	 it's sacrificed when it enters.)"
//
// Man-o'-War's enters bounce over any nonland permanent, anyone's, with
// Mulldrifter's evoke (CR 702.74a): cast for {1}{U}{U}, the creature still
// enters, so the bounce and the evoke sacrifice trigger together and their
// controller orders them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e0be003-4f42-4917-b85d-46baf2afd734",
		Name:         "Aethersnipe",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Evoke("{1}{U}{U}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Aethersnipe — return target nonland permanent to its owner's hand",
				effects.BounceTheTargetToHand), effects.TargetPermanent("target nonland permanent", effects.Nonland())),
		},
	})
}
