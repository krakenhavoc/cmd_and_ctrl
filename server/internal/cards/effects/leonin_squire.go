package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonin Squire — Creature — Cat Soldier {1}{W}, 2/2 (Fifth Dawn, #5604):
//
//	"When this creature enters, return target artifact card with mana
//	 value 1 or less from your graveyard to your hand."
//
// Auriok Salvagers' target, as an enters trigger. With no such card in the
// graveyard the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa19c9d7-fe93-4e93-a5ef-9af9c6cd1a5e",
		Name:         "Leonin Squire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Leonin Squire — return an artifact card with mana value 1 or less to your hand",
					returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("target artifact card with mana value 1 or less in your graveyard",
					Artifact(), ManaValueLE(1), YouOwn()),
			),
		},
	})
}
