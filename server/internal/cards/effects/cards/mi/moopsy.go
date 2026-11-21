package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moopsy — {G} Creature — Beast, 1/1 (Star Trek, #4973):
//
//	"Deathtouch
//	 {2}: This creature gains reach until end of turn."
//
// Frog Butler's reach activation without its mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "89b9efc8-d1cd-4b2a-acf3-4f0dd1629167",
		Name:            "Moopsy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gains reach until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("reach", "Moopsy — reach until end of turn"),
		}},
	})
}
