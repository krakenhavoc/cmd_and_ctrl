package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// High-Rise Sawjack (Streets of New Capenna, #5491):
//
//	Reach
//	Whenever this creature blocks a creature with flying, this creature gets +2/+0 until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0049438b-e37e-4c57-820e-f2ddf3322170",
		Name:            "High-Rise Sawjack",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.SkystingerBlockedAFlier,
				"High-Rise Sawjack — +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "High-Rise Sawjack — +2/+0 until end of turn")),
		},
	})
}
