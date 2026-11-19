package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skymarch Bloodletter — Creature — Vampire Soldier {2}{B}, 2/2
// (Foundations Jumpstart):
//
//	"Flying
//	 When this creature enters, target opponent loses 1 life and you gain
//	 1 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "41684707-5118-45cf-892d-46b66045d026",
		Name:            "Skymarch Bloodletter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Skymarch Bloodletter — target opponent loses 1 life and you gain 1 life", drainTargetOpponentOne),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
