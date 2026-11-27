package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slashing Tiger — Creature — Cat {2}{G}{G}, 3/3
// (Portal Three Kingdoms, #5622):
//
//	"Whenever this creature becomes blocked, it gets +2/+2 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c07a86de-ec29-478c-b6cb-057eb9d3e126",
		Name:         "Slashing Tiger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self, "Slashing Tiger — it gets +2/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Slashing Tiger — +2/+2 until end of turn")),
		},
	})
}
