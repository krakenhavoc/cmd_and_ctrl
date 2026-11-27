package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clergy en-Vec — Creature — Human Cleric {1}{W}, 1/1 (Tempest, S82 OLD b754,
// #5628):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this
//	 turn."
//
// Barrenton Medic's shield (PreventNextOneDamageToTheTarget, CR 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5a1b4c9-474a-42b7-b412-c2d43a3353db",
		Name:         "Clergy en-Vec",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextOneDamageToTheTarget("Clergy en-Vec"),
		}},
	})
}
