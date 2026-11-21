package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graceful Cat — Creature — Cat {2}{W}, 2/2
// (Amonkhet, #5535):
//
//	"Whenever this creature attacks, it gets +1/+1 until end of turn."
//
// Soltari Trooper's attack trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69fffad0-6469-4623-9977-347143fa7289",
		Name:         "Graceful Cat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Graceful Cat — it gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Graceful Cat — +1/+1")),
		},
	})
}
