package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Razorclaw Bear — Creature — Bear {2}{G}{G}, 3/3 (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Whenever this creature becomes blocked, it gets +2/+2 until end of turn."
//
// Triggers once as it becomes blocked, however many creatures block it (CR 509.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ee07899-e6d6-4dcf-8596-4d2420690d7c",
		Name:         "Razorclaw Bear",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self,
				"Razorclaw Bear — it gets +2/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Razorclaw Bear — +2/+2")),
		},
	})
}
