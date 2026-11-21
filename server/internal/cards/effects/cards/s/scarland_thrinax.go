package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scarland Thrinax — Creature — Lizard {B}{R}{G}, 2/2 (Conflux, #5587):
//
//	"Sacrifice a creature: Put a +1/+1 counter on this creature."
//
// Bloodflow Connoisseur's row: the sacrifice is the cost, so it may be
// the Thrinax itself (the counter then has nowhere to go), and with no
// mana or {T} it can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a185f90-69eb-496b-8e38-8dcf241a84a4",
		Name:         "Scarland Thrinax",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Effect:  effects.PutACounterOnThis(game.CounterPlusOne),
		}},
	})
}
