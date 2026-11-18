package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snarespinner — Creature — Spider {1}{G}, 1/3:
//
//	"Reach
//	 Whenever this creature blocks a creature with flying, this creature
//	 gets +2/+0 until end of turn."
//
// The block trigger is Skystinger's (skystingerBlockedAFlier): the
// blocked attacker's flying is read when the block is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0f63de7b-03f2-49ad-b217-88fe9c50d68a",
		Name:            "Snarespinner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, skystingerBlockedAFlier,
				"Snarespinner — +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Snarespinner — +2/+0 until end of turn")),
		},
	})
}
