package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcano Imp — Creature — Imp {3}{B}, 2/2 (Planeshift, S82 OLD #5615):
//
//	"Flying
//	 {1}{R}: This creature gains first strike until end of turn."
//
// Boreal Griffin's first-strike grant through the shared
// ThisGainsKeywordUntilEndOfTurn body, paid with {1}{R}. Gained after blockers
// are declared, it still deals its damage in the first-strike step (CR 510.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d2b1b4dd-e860-40d8-8e07-1dee5be65bc4",
		Name:            "Volcano Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Volcano Imp — first strike until end of turn"),
		}},
	})
}
