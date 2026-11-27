package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orim, Samite Healer — Legendary Creature — Human Cleric {1}{W}{W}, 1/3
// (Tempest, S82 OLD b754, #5628):
//
//	"{T}: Prevent the next 3 damage that would be dealt to any target this
//	 turn."
//
// Barrenton Medic's shield (PreventNextDamageToTheTarget, CR 615.7) at 3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d405e5b-0cee-4c8f-b3ef-65765853b8a4",
		Name:         "Orim, Samite Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 3 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextDamageToTheTarget("Orim, Samite Healer", 3),
		}},
	})
}
