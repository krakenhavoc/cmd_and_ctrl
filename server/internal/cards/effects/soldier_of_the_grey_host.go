package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soldier of the Grey Host — Creature — Spirit Soldier {3}{W}, 2/2 (The Hobbit):
//
//	"Flash
//	 Flying
//	 When this creature enters, target creature gets +2/+0 until end of
//	 turn."
//
// Any creature, the Soldier included, picked as the trigger goes on the
// stack (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a04c6059-d9e2-4b2b-9db5-f936d69da4d4",
		Name:            "Soldier of the Grey Host",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Soldier of the Grey Host — target creature gets +2/+0 until end of turn",
				boostTargetUntilEOT("Soldier of the Grey Host — +2/+0", 2, 0)),
				TargetCreature("target creature")),
		},
	})
}
