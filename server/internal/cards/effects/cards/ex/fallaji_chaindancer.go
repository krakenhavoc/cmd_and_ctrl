package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fallaji Chaindancer — Creature — Human Soldier {3}{R}, 2/4 (The Brothers' War, #5482):
//
//	"{2}: This creature gains double strike until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb690399-0fb8-4bd8-a73d-5e0e594ff2ae",
		Name:         "Fallaji Chaindancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gains double strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("double strike", "Fallaji Chaindancer — double strike until end of turn"),
		}},
	})
}
