package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rotcrown Ghoul — {4}{U} Creature — Zombie 3/3 (Avacyn Restored, #5570):
//
//	"When this creature dies, target player mills five cards."
//
// A dies trigger aimed at a player, Drownyard Amalgam's mill for
// five.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9a420d4-a834-4a14-b874-1db96624d8ef",
		Name:         "Rotcrown Ghoul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Rotcrown Ghoul — target player mills five cards", targetPlayerMills(5)), TargetPlayer("target player")),
		},
	})
}
