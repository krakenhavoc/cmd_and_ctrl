package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gang of Elk — Creature — Elk Beast {5}{G}, 5/4 (Urza's Legacy, S82 OLD b715, #5623):
//
//	"Whenever this creature becomes blocked, it gets +2/+2 until end of turn for each creature blocking it."
//
// Berserk Murlodont's becomes-blocked trigger on itself: once per
// attacker however many blockers (CR 509.1h), the blockers counted as the
// trigger resolves (ThisGetsForEachBlocker), the bonus fixed then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b99d6d51-3ead-4b81-8e0c-f458619ebb24",
		Name:         "Gang of Elk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.ThisBecomesBlocked,
				"Gang of Elk — it gets +2/+2 until end of turn for each creature blocking it",
				effects.ThisGetsForEachBlocker(2, 2, "Gang of Elk — +2/+2 for each creature blocking it")),
		},
	})
}
