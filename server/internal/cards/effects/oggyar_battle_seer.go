package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oggyar Battle-Seer — Creature — Ogre Shaman {3}{U}{R}, 3/4 (Strixhaven:
// School of Mages, #5502):
//
//	"Haste
//	 {T}: Scry 1."
//
// Haste is the printed keyword, so the scry row can be used the turn it
// arrives; the row is Sigiled Starfish's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "645005dd-3438-4b88-bb19-561f9679a022",
		Name:            "Oggyar Battle-Seer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
