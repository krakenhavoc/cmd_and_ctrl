package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Putrid Leech — Creature — Zombie Leech {B}{G}, 2/2 (Alara Reborn,
// #5586):
//
//	"Pay 2 life: This creature gets +2/+2 until end of turn. Activate only
//	 once each turn."
//
// Desolation Prowler's row: a life cost, a self pump that ends in the
// cleanup step, and the once-each-turn activation limit (CR 602.5b).
//
// No simplification.
func init() {
	const label = "Pay 2 life: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "16b723bc-aab9-40f2-8e1b-b9c2bd2af3bc",
		Name:         "Putrid Leech",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.PayLife(2),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Putrid Leech — +2/+2 until end of turn"),
		}},
	})
}
