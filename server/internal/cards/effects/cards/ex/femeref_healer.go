package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Femeref Healer — Creature — Human Cleric {1}{W}, 1/1 (Mirage, S82 OLD b808):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this
//	 turn."
//
// Abuna Acolyte's first ability: a charged prevention shield (CR 615.7) on the
// target, which lasts until the turn ends.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2117373-6f0a-41c8-ab6a-9014b9628e64",
		Name:         "Femeref Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextOneDamageToTheTarget("Femeref Healer"),
		}},
	})
}
