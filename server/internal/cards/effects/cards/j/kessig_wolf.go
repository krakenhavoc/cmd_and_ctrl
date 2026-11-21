package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kessig Wolf — Creature — Wolf {2}{R}, 3/1 (Innistrad, #5572):
//
//	"{1}{R}: This creature gains first strike until end of turn."
//
// Burning Shield Askari's first-strike grant at {1}{R}, until end of
// turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b82045f-5a2b-4e6e-976c-986c80ecd8b2",
		Name:         "Kessig Wolf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Kessig Wolf — first strike until end of turn"),
		}},
	})
}
