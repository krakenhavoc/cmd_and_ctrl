package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fume Spitter — Creature — Phyrexian Horror {B}, 1/1 (Scars of Mirrodin,
// #5577):
//
//	"Sacrifice this creature: Put a -1/-1 counter on target creature."
//
// The sacrifice is the cost, paid as it is activated (CR 602.2b), so the
// counter still goes on if the Spitter is gone; a target gone in response
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6556a15b-be4b-45ea-867d-3a8e26202b42",
		Name:         "Fume Spitter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Put a -1/-1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeThis(),
			Targets: effects.Harmful(effects.TargetCreature("target creature")),
			Effect:  effects.PutACounterOnTheTarget(game.CounterMinusOne),
		}},
	})
}
