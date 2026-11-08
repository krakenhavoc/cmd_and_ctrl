package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unburied Earthcarver — Creature — Human Warrior {1}{B}, 2/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"{2}, Sacrifice another creature: Put a +1/+1 counter on this
//	 creature."
//
// Boneyard Desecrator's sacrifice outlet without the outlaw rider: the
// other creature is sacrificed as the ability is activated (the
// Earthcarver itself is not a legal payment), and the counter goes on the
// Earthcarver if it is still the same object on the battlefield when the
// ability resolves (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75da657e-0e36-4f2a-9b81-b7eba7d8e050",
		Name:         "Unburied Earthcarver",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice another creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 1},
			Cost:    Plus(ManaCost("{2}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect:  putACounterOnThis(game.CounterPlusOne),
		}},
	})
}
