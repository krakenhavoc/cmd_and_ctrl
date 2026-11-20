package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ac193909-7e81-4082-ad02-2a60d142a0ab",
		Name:         "Blood Host",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature: Put a +1/+1 counter on this creature and you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 1, LifeGain: 2},
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := putACounterOnThis(game.CounterPlusOne)(g, item); err != nil {
					return err
				}
				return GainLife{Amount: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
