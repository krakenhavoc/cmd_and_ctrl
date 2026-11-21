package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadbridge Shaman — Creature — Elf Shaman {2}{B}, 3/1 (Magic Origins, #5549):
//
//	"When this creature dies, target opponent discards a card."
//
// Ravenous Rats' discard on a death trigger: the opponent is targeted as
// the trigger goes on the stack and picks their own discard (CR 701.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a57389ad-7f2a-47a6-8ae6-2821503cbb28",
		Name:         "Deadbridge Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Deadbridge Shaman — target opponent discards a card", effects.TargetOpponentDiscards(1)),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
