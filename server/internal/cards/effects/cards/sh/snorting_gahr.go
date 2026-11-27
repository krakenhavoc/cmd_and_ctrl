package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Snorting Gahr — Creature — Rhino Beast {2}{G}{G}, 3/3 (Mercadian Masques,
// S82 OLD, #5619):
//
//	"Whenever this creature becomes blocked, it gets +2/+2 until end of turn."
//
// "Becomes blocked" is once per blocked attacker however many creatures block it
// (CR 509.3c), and the bonus is the Gahr's own, only while it is still the same
// object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e7f67d6-db12-4034-b8e3-9c9d401c8f57",
		Name:         "Snorting Gahr",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self, "Snorting Gahr — it gets +2/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Snorting Gahr — +2/+2 until end of turn")),
		},
	})
}
