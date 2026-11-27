package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctum Custodian — Creature — Human Cleric {2}{W} 1/2 (Urza's Saga, #5624):
//
//	"{T}: Prevent the next 2 damage that would be dealt to any target this
//	 turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d979fbf-5009-47a1-8d77-b1ce80bbe90d",
		Name:         "Sanctum Custodian",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 2 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextDamageToTheTarget("Sanctum Custodian", 2),
		}},
	})
}
