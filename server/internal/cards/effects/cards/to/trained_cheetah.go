package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trained Cheetah — Creature — Cat {2}{G}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"Whenever this creature becomes blocked, it gets +1/+1 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dad3b459-725e-4bf5-bf7a-bd89866ea8c4",
		Name:         "Trained Cheetah",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self, "Trained Cheetah — it gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Trained Cheetah — +1/+1 until end of turn")),
		},
	})
}
