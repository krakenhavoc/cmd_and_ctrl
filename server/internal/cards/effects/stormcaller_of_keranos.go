package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "16a66bbe-90f6-4c29-a69d-6ff254082956",
		Name:            "Stormcaller of Keranos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    ManaCost("{1}{U}"),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
