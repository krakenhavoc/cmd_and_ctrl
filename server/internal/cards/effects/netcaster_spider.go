package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Netcaster Spider — Creature — Spider {2}{G}, 2/3 (Magic 2015, #5556):
//
//	"Reach
//	 Whenever this creature blocks a creature with flying, this creature
//	 gets +2/+0 until end of turn."
//
// Snarespinner's block trigger (skystingerBlockedAFlier): the blocked
// attacker's flying is read when the block is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0ddb3fa0-447c-40c5-928a-541f3d543014",
		Name:            "Netcaster Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, skystingerBlockedAFlier,
				"Netcaster Spider — +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Netcaster Spider — +2/+0 until end of turn")),
		},
	})
}
