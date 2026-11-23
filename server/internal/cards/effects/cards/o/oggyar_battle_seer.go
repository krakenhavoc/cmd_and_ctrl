package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "645005dd-3438-4b88-bb19-561f9679a022",
		Name:            "Oggyar Battle-Seer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Scry{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
