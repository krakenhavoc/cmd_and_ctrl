package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gutless Ghoul — Snow Creature — Zombie {2}{B}, 2/2 (Coldsnap, #5597):
//
//	"{1}, Sacrifice a creature: You gain 2 life."
//
// A sacrifice outlet (Miren's cost, whose creature may be this Ghoul itself)
// with a fixed gain of 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ebc166d-98c5-4ab7-9610-bc5bcd231d80",
		Name:         "Gutless Ghoul",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: You gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 2},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
