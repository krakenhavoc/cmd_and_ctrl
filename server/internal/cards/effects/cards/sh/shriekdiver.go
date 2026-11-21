package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shriekdiver — Creature — Zombie Bird Warrior {2}{B}, 2/1 (War of the Spark, #5520):
//
//	"Flying
//	 {1}: This creature gains haste until end of turn."
//
// Flying is printed data; the haste is a layer-6 grant on itself until
// end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f51e1660-e1c8-4122-8c43-321c820a1fc1",
		Name:            "Shriekdiver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This creature gains haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("haste", "Shriekdiver — haste until end of turn"),
		}},
	})
}
