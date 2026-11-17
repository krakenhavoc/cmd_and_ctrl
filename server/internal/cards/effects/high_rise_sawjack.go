package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// High-Rise Sawjack (Streets of New Capenna, #5491):
//
//	Reach
//	Whenever this creature blocks a creature with flying, this creature gets +2/+0 until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0049438b-e37e-4c57-820e-f2ddf3322170",
		Name:            "High-Rise Sawjack",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, skystingerBlockedAFlier,
				"High-Rise Sawjack — +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "High-Rise Sawjack — +2/+0 until end of turn")),
		},
	})
}
