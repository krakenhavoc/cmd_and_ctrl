package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladed Ambassador — Creature — Phyrexian Soldier {1}{W}, 3/1 (Phyrexia: All
// Will Be One):
//
//	"This creature enters with an oil counter on it.
//	 {1}, Remove an oil counter from this creature: This creature gains
//	 indestructible until end of turn."
//
// The oil counter is a CR 614.1c entry replacement; removing it is a cost
// paid at announcement, so an Ambassador with none cannot activate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3088ee75-3b51-4cbc-8987-76cb93f416b1",
		Name:         "Bladed Ambassador",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 1, "Bladed Ambassador: enters with an oil counter on it")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Remove an oil counter from this creature: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis(effects.OilCounter, 1)),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Bladed Ambassador — indestructible until end of turn"),
		}},
	})
}
