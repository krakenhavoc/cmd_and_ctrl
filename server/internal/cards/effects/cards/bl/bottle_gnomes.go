package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bottle Gnomes — Artifact Creature — Gnome {3} 1/3 (Commander 2014):
//
//	"Sacrifice this creature: You gain 3 life."
//
// The sacrifice is the whole cost, paid on activation, so removal in
// response cannot stop the life gain.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54b5e429-7a44-480d-bea4-4f8eeb7449b5",
		Name:         "Bottle Gnomes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Sacrifice this creature: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.SacrificeThis(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
