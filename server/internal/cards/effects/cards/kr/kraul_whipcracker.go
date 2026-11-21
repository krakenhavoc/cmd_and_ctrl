package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b563ae31-338e-4bcc-a05f-d7ed7c3fe46e",
		Name:            "Kraul Whipcracker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Kraul Whipcracker — destroy target token an opponent controls", effects.DestroyFirstLegalTarget),
				effects.TargetPermanent("target token an opponent controls", effects.IsTokenPredicate(), effects.OpponentControls())),
		},
	})
}
