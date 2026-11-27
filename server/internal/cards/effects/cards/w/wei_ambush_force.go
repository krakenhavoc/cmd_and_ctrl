package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wei Ambush Force — Creature — Human Soldier {1}{B}, 1/1
// (Portal Three Kingdoms, #5622):
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d1446f4-cf71-4798-b921-8f6824350319",
		Name:         "Wei Ambush Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Wei Ambush Force — it gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Wei Ambush Force — +2/+0 until end of turn")),
		},
	})
}
