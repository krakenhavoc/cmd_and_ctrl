package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tivadar of Thorn — Legendary Creature — Human Knight {1}{W}{W}, 2/2
// (Time Spiral, #5596):
//
//	"First strike, protection from red
//	 When Tivadar enters, destroy target Goblin."
//
// Avalanche Riders' enters trigger aimed at a Goblin: any Goblin
// permanent, whoever controls it, and the trigger is removed if there is
// none to target (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c2c09308-988b-4d71-83c8-e2a2dc19a0d9",
		Name:            "Tivadar of Thorn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "protection from red"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Tivadar of Thorn — destroy target Goblin", destroyFirstLegalCardTarget),
				TargetPermanent("target Goblin", Subtype("Goblin"))),
		},
	})
}
