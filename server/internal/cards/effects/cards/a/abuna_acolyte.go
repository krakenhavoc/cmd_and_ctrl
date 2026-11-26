package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abuna Acolyte — Creature — Cat Cleric {1}{W}, 1/1 (Scars of Mirrodin,
// #5577):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this
//	 turn.
//	 {T}: Prevent the next 2 damage that would be dealt to target artifact
//	 creature this turn."
//
// Two separate {T} abilities, each a charged prevention shield (CR 615.7) on
// its own target: any target, and an artifact creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf082161-47ff-46b9-926d-3fa5e5727861",
		Name:         "Abuna Acolyte",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.PreventNextOneDamageToTheTarget("Abuna Acolyte"),
			},
			{
				Label:   "{T}: Prevent the next 2 damage that would be dealt to target artifact creature this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
				Effect:  effects.PreventNextDamageToTheTarget("Abuna Acolyte", 2),
			},
		},
	})
}
