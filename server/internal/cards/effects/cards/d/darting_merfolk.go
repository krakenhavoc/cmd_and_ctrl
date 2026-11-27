package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Darting Merfolk — Creature — Merfolk {1}{U}, 1/1 (Mercadian Masques, S82
// OLD b681, #5619):
//
//	"{U}: Return this creature to its owner's hand."
//
// Cavern Harpy's self-bounce on a {U} cost, at instant speed as often as it
// can be paid. A Merfolk gone by resolution stays where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f33e684-a4e4-4528-9a20-3df902bc2cde",
		Name:         "Darting Merfolk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
