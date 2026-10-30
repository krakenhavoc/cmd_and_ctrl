package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hammerhead, Maggia Boss — Legendary Creature — Human Rogue Villain
// {1}{B}, 2/1 (Marvel Super Heroes Commander, #4455):
//
//	"Sacrifice another creature or artifact: Put a +1/+1 counter on
//	 Hammerhead."
//
// The cost is Umbral Collar Zealot's (SacrificeOther with an Another
// creature-or-artifact spec); the counter goes on Hammerhead.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ff5a5ed-d6db-4dfa-b1fd-38ad872ad30a",
		Name:         "Hammerhead, Maggia Boss",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another creature or artifact: Put a +1/+1 counter on Hammerhead.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				SacrificeOther: Another(sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()))),
			},
			Effect: Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
