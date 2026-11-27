package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shock Troops — Creature — Human Soldier {3}{R}, 2/2 (Eighth Edition, S82 OLD
// batch 611, #5607):
//
//	"Sacrifice this creature: It deals 2 damage to any target."
//
// Mogg Fanatic's ability at 2: the sacrifice is the cost, so the damage is
// dealt by the creature as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdf2a8da-2933-44c5-b483-f035144da752",
		Name:         "Shock Troops",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: It deals 2 damage to any target.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetAny(),
			Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
			Effect:  effects.B33DamageChosenTargetFromSource(2),
		}},
	})
}
