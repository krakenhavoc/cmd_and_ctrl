package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whiptongue Frog — Creature — Frog {2}{U}, 1/3 (Exodus, #5626):
//
//	"{U}: This creature gains flying until end of turn."
//
// Killer Whale's row on a smaller body: the shared
// thisGainsKeywordUntilEndOfTurn grant, ending in the cleanup step
// (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edc961bd-5ec9-47ed-9d34-e26a48c13702",
		Name:         "Whiptongue Frog",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Whiptongue Frog — flying"),
		}},
	})
}
