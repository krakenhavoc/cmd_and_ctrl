package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Graceful Cat — Creature — Cat {2}{W}, 2/2
// (Amonkhet, #5535):
//
//	"Whenever this creature attacks, it gets +1/+1 until end of turn."
//
// Soltari Trooper's attack trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69fffad0-6469-4623-9977-347143fa7289",
		Name:         "Graceful Cat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Graceful Cat — it gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Graceful Cat — +1/+1")),
		},
	})
}
