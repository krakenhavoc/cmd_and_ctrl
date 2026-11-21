package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2ff5a5ed-d6db-4dfa-b1fd-38ad872ad30a",
		Name:         "Hammerhead, Maggia Boss",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature or artifact: Put a +1/+1 counter on Hammerhead.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or artifact",
					effects.Or(effects.Creature(), effects.Artifact()))),
			},
			Effect: effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
