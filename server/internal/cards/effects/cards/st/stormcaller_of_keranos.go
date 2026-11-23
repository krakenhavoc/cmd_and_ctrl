package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormcaller of Keranos — Creature — Human Shaman {2}{R}, 2/2 (Born of the Gods,
// #5559):
//
//	"Haste
//	 {1}{U}: Scry 1. (Look at the top card of your library. You may put
//	 that card on the bottom.)"
//
// Tidepool Turtle's ability at {1}{U}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "16a66bbe-90f6-4c29-a69d-6ff254082956",
		Name:            "Stormcaller of Keranos",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
