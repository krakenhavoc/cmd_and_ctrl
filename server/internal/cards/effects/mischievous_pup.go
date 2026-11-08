package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mischievous Pup — Creature — Dog {2}{W}, 3/1 (Foundations, #4465):
//
//	"Flash
//	 When this creature enters, return up to one other target permanent
//	 you control to its owner's hand."
//
// Mirkwood Nurturer's target clause without its counter, shared with
// Stickytongue Sentinel as whenThisEntersReturnAnotherPermanentYouControl:
// the still-legal pick returns to its owner's hand (bounceTheTargetToHand),
// and choosing none returns nothing. Flash rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e75b187d-0e3b-4d94-a940-8b7e4e7ed1ca",
		Name:            "Mischievous Pup",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			whenThisEntersReturnAnotherPermanentYouControl("Mischievous Pup"),
		},
	})
}
