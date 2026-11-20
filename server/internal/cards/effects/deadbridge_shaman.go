package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadbridge Shaman — Creature — Elf Shaman {2}{B}, 3/1 (Magic Origins, #5549):
//
//	"When this creature dies, target opponent discards a card."
//
// Ravenous Rats' discard on a death trigger: the opponent is targeted as
// the trigger goes on the stack and picks their own discard (CR 701.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a57389ad-7f2a-47a6-8ae6-2821503cbb28",
		Name:         "Deadbridge Shaman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Deadbridge Shaman — target opponent discards a card", targetOpponentDiscards(1)),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
