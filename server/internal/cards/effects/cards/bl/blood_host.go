package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Host — Creature — Vampire {3}{B}{B}, 3/3 (Magic 2015, #5556):
//
//	"{1}{B}, Sacrifice another creature: Put a +1/+1 counter on this
//	 creature and you gain 2 life."
//
// Hungry Ghoul's sacrifice outlet at {1}{B} with a life rider: the other
// creature is sacrificed as the ability is activated (the Host itself is
// not a legal payment), the counter goes on the Host if it is still on
// the battlefield, and the 2 life is gained either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac193909-7e81-4082-ad02-2a60d142a0ab",
		Name:         "Blood Host",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature: Put a +1/+1 counter on this creature and you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 1, LifeGain: 2},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.PutACounterOnThis(game.CounterPlusOne)(g, item); err != nil {
					return err
				}
				return effects.GainLife{Amount: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
