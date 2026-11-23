package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8c17f48b-467e-4953-85d0-221cef5b3bff",
		Name:         "Lurking Nightstalker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Lurking Nightstalker — it gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Lurking Nightstalker — +2/+0")),
		},
	})
}
