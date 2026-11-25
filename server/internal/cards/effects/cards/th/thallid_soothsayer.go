package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thallid Soothsayer — Creature — Fungus {3}{B}, 2/3 (Game Night):
//
//	"{2}, Sacrifice a creature: Draw a card."
//
// Grim Backwoods' sacrifice-outlet draw on a creature: the creature is paid
// at announce, so its dies-triggers land above the ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eee801ea-e2dc-4b03-a9f9-e87566cb2ae8",
		Name:         "Thallid Soothsayer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice a creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
