package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blighted Bat — Creature — Zombie Bat {2}{B}, 2/1 (Jumpstart, #5510):
//
//	"Flying
//	 {1}: This creature gains haste until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "542887e9-1032-4396-9cbb-aa3c11ee43a0",
		Name:            "Blighted Bat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This creature gains haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("haste", "Blighted Bat — gains haste until end of turn"),
		}},
	})
}
