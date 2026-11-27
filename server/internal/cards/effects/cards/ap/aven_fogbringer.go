package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Fogbringer — Creature — Bird Wizard {3}{U}, 2/1 (Judgment):
//
//	"Flying
//	 When this creature enters, return target land to its owner's hand."
//
// Riftwing Cloudskate's mandatory enters bounce, aimed at a land of any
// player's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3fb3d4b3-1034-449c-864b-633e7f39d73f",
		Name:            "Aven Fogbringer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Aven Fogbringer — return target land to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetPermanent("target land", effects.Land())),
		},
	})
}
