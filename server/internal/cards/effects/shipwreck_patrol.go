package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shipwreck Patrol — Creature — Merfolk Warrior {3}{U}, 3/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"When this creature enters, tap target creature an opponent controls and
//	 put a stun counter on it."
//
// Cryotheory Adept's tap-and-stun body (tapTargetAndStunIt).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9851f207-9f75-4ceb-8024-8f9bf21f90ad",
		Name:         "Shipwreck Patrol",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Shipwreck Patrol — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
