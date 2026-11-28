package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alabaster Wall — Creature — Wall {2}{W}, 0/4 (Mercadian Masques, S82 OLD
// b682, #5619):
//
//	"Defender
//	 {T}: Prevent the next 1 damage that would be dealt to any target this
//	 turn."
//
// Defender is a printed keyword; the shield is Abuna Acolyte's
// (PreventNextOneDamageToTheTarget, CR 615.7) on any target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f9f52651-cc64-48c5-bc18-cd5f646dfb1f",
		Name:            "Alabaster Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextOneDamageToTheTarget("Alabaster Wall"),
		}},
	})
}
