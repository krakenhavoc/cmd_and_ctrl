package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carnage Altar — Artifact {2} (Commander 2013, S82 OLD b380, #5560):
//
//	"{3}, Sacrifice a creature: Draw a card."
//
// Thallid Soothsayer's sacrifice-and-draw at {3}: the creature is a cost, any
// creature you control (the Altar is no creature), and the draw is one card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa05900f-0f04-407e-931c-fea8f91e78e3",
		Name:         "Carnage Altar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, Sacrifice a creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
