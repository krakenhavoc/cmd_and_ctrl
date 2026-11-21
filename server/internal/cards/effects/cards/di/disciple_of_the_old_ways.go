package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disciple of the Old Ways — Creature — Human Warrior {1}{G}, 2/2 (Gatecrash, #5565):
//
//	"{R}: This creature gains first strike until end of turn."
//
// Burning Shield Askari's grant at a different cost, pinned to this
// creature until cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d19d69c5-9f11-4ac1-9fc7-1636689f23a0",
		Name:         "Disciple of the Old Ways",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Disciple of the Old Ways — first strike until end of turn"),
		}},
	})
}
