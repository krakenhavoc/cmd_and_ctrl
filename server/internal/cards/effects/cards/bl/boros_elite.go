package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boros Elite — Creature — Human Soldier {W}, 1/1 (Gatecrash, #5565):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, this creature gets +2/+2 until end of turn."
//
// Battalion is Boros Strike-Captain's group test (gtcBattalion); the
// bonus is pinned to this creature and ends in the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c83fb45-03af-4f1a-b7e3-2c4f47511658",
		Name:         "Boros Elite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GtcBattalion("Boros Elite — it gets +2/+2 until end of turn (battalion)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Boros Elite — +2/+2")),
		},
	})
}
