package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hungry Ghoul — Creature — Zombie {1}{B}, 2/2 (Foundations, #4465):
//
//	"{1}, Sacrifice another creature: Put a +1/+1 counter on this
//	 creature."
//
// Unburied Earthcarver's sacrifice outlet for {1}: the other creature is
// sacrificed as the ability is activated (the Ghoul itself is not a legal
// payment), and the counter goes on the Ghoul if it is still on the
// battlefield when the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a7a55a3-568c-4c31-9460-5b94dc0121d8",
		Name:         "Hungry Ghoul",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 1},
			Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect:  putACounterOnThis(game.CounterPlusOne),
		}},
	})
}
