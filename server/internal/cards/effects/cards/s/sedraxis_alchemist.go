package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sedraxis Alchemist — Creature — Zombie Wizard {2}{B}, 2/2 (Conflux,
// #5587):
//
//	"When this creature enters, if you control a blue permanent, return
//	 target nonland permanent to its owner's hand."
//
// An intervening-if enters trigger (CR 603.4,
// conB2WhenThisEntersIfYouControlA) whose body is Aethersnipe's bounce.
// The target is mandatory, so with a blue permanent the Alchemist may
// have to return a permanent of your own (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56f3f1b4-c36a-41f3-95a1-45e9d7e74cf9",
		Name:         "Sedraxis Alchemist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ConB2WhenThisEntersIfYouControlA("U", "blue",
				"Sedraxis Alchemist — return target nonland permanent to its owner's hand",
				effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				effects.BounceTheTargetToHand),
		},
	})
}
