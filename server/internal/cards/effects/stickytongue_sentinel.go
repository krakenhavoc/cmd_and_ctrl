package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stickytongue Sentinel — Creature — Frog Warrior {2}{G}, 3/3
// (Bloomburrow, #4467):
//
//	"Reach
//	 When this creature enters, return up to one other target permanent
//	 you control to its owner's hand."
//
// Mischievous Pup's entry trigger, shared as
// whenThisEntersReturnAnotherPermanentYouControl: the still-legal pick
// returns to its owner's hand, and choosing none returns nothing. Reach
// rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "899f0931-4e9b-4fa8-9f6d-d13a7de05334",
		Name:            "Stickytongue Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			whenThisEntersReturnAnotherPermanentYouControl("Stickytongue Sentinel"),
		},
	})
}
