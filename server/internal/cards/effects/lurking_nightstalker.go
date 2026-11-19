package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lurking Nightstalker — Creature — Nightstalker {B}{B}, 1/1 (Portal
// Second Age, #5625):
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// Vicious Kavu's attack trigger: the boost is pinned to the Nightstalker
// itself, so one that has left the battlefield before the trigger
// resolves gets nothing, and it ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c17f48b-467e-4953-85d0-221cef5b3bff",
		Name:         "Lurking Nightstalker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Lurking Nightstalker — it gets +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Lurking Nightstalker — +2/+0")),
		},
	})
}
