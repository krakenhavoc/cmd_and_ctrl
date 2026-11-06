package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kraul Whipcracker — Creature — Insect Assassin {B}{G}, 3/2 (Murders at
// Karlov Manor, #4469):
//
//	"Reach
//	 When this creature enters, destroy target token an opponent
//	 controls."
//
// The enters trigger targets as it goes on the stack; with no opposing
// token it is removed for want of a target. A token that left in response
// is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b563ae31-338e-4bcc-a05f-d7ed7c3fe46e",
		Name:            "Kraul Whipcracker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Kraul Whipcracker — destroy target token an opponent controls", destroyFirstLegalTarget),
				TargetPermanent("target token an opponent controls", IsTokenPredicate(), OpponentControls())),
		},
	})
}
