package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warmind Infantry — Creature — Elemental Soldier {2}{R}, 2/3 (Gatecrash, #5565):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, this creature gets +2/+0 until end of turn."
//
// Battalion is Boros Strike-Captain's group test (gtcBattalion); the
// bonus is pinned to this creature and ends in the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e32750a-8e5f-4d81-92a8-c3d4ea1477fd",
		Name:         "Warmind Infantry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GtcBattalion("Warmind Infantry — it gets +2/+0 until end of turn (battalion)",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Warmind Infantry — +2/+0")),
		},
	})
}
