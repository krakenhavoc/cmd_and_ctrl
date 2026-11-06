package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desperate Bloodseeker — Creature — Vampire {1}{B}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"Lifelink
//	 When this creature enters, target player mills two cards. (They put
//	 the top two cards of their library into their graveyard.)"
//
// Lifelink is the printed keyword. The player is the enters trigger's target,
// chosen as it goes on the stack and checked again on resolution (CR 608.2b);
// you may name yourself to fill your own graveyard. The mill is Scrabbling
// Skullcrab's body (targetPlayerMills, CR 701.17a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "de8d7cad-02c7-43f1-896e-eaa73f016ada",
		Name:            "Desperate Bloodseeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Desperate Bloodseeker — target player mills two cards", targetPlayerMills(2)), TargetPlayer("target player")),
		},
	})
}
