package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vicious Kavu — Creature — Kavu {1}{B}{R}, 2/2 (Invasion, #5616):
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// Flowstone Charger's attack trigger at +2/+0: the boost is pinned to
// the Kavu itself, so a Kavu that has left the battlefield before the
// trigger resolves gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "236fbf12-f908-4918-babb-1558e767cbe3",
		Name:         "Vicious Kavu",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Vicious Kavu — it gets +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Vicious Kavu — +2/+0")),
		},
	})
}
