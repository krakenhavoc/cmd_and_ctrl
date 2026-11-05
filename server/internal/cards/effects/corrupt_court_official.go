package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corrupt Court Official — Creature — Human Advisor {1}{B}, 1/1:
//
//	"When this creature enters, target opponent discards a card."
//
// HYDRA Infiltration's entry with one card: the opponent is targeted as
// the trigger goes on the stack and chooses their own discard (CR
// 701.9a). An opponent who has left by resolution is an illegal target
// and nothing happens (CR 608.2b); an empty hand discards nothing.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "ec84201d-757f-4a7a-b3a2-85ccd5ee81ae",
		Name:         "Corrupt Court Official",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Corrupt Court Official — target opponent discards a card", targetOpponentDiscards(1)),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
