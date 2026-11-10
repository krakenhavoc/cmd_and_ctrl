package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exosuit Savior — Creature — Human Soldier {2}{W}, 2/2 (Edge of
// Eternities, #4461):
//
//	"Flying
//	 When this creature enters, return up to one other target permanent
//	 you control to its owner's hand."
//
// Mischievous Pup's enters trigger in a flying body: the still-legal
// pick returns to its owner's hand (bounceTheTargetToHand), and choosing
// none returns nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8b13ec4d-812a-48b0-8a64-6847a694a03a",
		Name:            "Exosuit Savior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Exosuit Savior — return up to one other target permanent you control to its owner's hand",
					bounceTheTargetToHand),
				Another(TargetPermanent("up to one other target permanent you control", YouControl())).WithCount(0, 1),
			),
		},
	})
}
