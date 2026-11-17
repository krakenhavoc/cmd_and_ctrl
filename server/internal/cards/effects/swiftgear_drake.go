package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swiftgear Drake — Artifact Creature — Drake {5}, 2/4 (The Brothers' War,
// #5482):
//
//	"Flying, haste
//	 When this creature enters, put up to one target card from a graveyard
//	 on the bottom of its owner's library."
//
// The Recluse's tuck on an enter trigger; "up to one", any graveyard, and
// the Drake's own graveyard is just another graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b7618312-6835-4851-9651-b5402001b3b0",
		Name:            "Swiftgear Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Swiftgear Drake — put up to one target card from a graveyard on the bottom of its owner's library",
					socB37PutTargetGraveyardCardOnBottom),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1),
			),
		},
	})
}
