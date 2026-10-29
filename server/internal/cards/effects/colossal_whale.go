package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossal Whale — Creature — Whale {5}{U}{U}, 5/5 (The Hobbit):
//
//	"Islandwalk
//	 Whenever this creature attacks, you may exile target creature defending
//	 player controls until this creature leaves the battlefield."
//
// Islandwalk is a printed keyword. The attack trigger is Coveted Peacock's
// shape for its target (TargetCreatureDefendingPlayerControls reads the
// defending player off this attack) with the "you may" asked as it resolves
// (CR 603.5), and the exile is the Oblivion Ring family's
// (exileChosenTargetUntilThisLeaves): if the Whale has already left by then
// nothing is exiled (CR 610.3c), and if it leaves later the creature returns
// under its owner's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cd04b735-125f-47b5-a68e-ab2a008100b9",
		Name:            "Colossal Whale",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches:     []game.EventKind{game.EventAttack},
				AppliesTo:   ThisAttacked,
				TargetsFrom: TargetCreatureDefendingPlayerControls,
				Key:         "Colossal Whale — exile target creature defending player controls until this creature leaves the battlefield",
				Effect: exileChosenTargetUntilThisLeaves(
					"Colossal Whale — the exiled creature returns when Colossal Whale leaves the battlefield"),
			}, "Colossal Whale — exile target creature defending player controls?"),
		},
	})
}
