package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Returned Centaur — Creature — Zombie Centaur {3}{B}, 2/4 (Theros, #5561):
//
//	"When this creature enters, target player mills four cards."
//
// Wall of Lost Thoughts' trigger: any player may be the target, its
// controller included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "223c97fb-dc96-41ba-be9d-bad9b5e4876f",
		Name:         "Returned Centaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Returned Centaur — target player mills four cards", TargetPlayerMills(4)),
				TargetPlayer("target player")),
		},
	})
}
