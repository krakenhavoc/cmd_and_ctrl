package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moorland Inquisitor — {1}{W} Creature — Human Soldier 2/2 (Avacyn Restored, #5570):
//
//	"{2}{W}: This creature gains first strike until end of turn. (It
//	 deals combat damage before creatures without first strike.)"
//
// Noble Panther's first-strike row at {2}{W}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76bbd280-dcbc-43b3-a8a1-603443b2ccaa",
		Name:         "Moorland Inquisitor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    effects.ManaCost("{2}{W}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Moorland Inquisitor — first strike until end of turn"),
		}},
	})
}
