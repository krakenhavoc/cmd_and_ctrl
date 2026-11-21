package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Harpy — Creature — Harpy {2}{B}, 1/2 (Core Set 2019, #5527):
//
//	"Flying
//	{1}, Sacrifice another creature: Put a +1/+1 counter on this creature."
//
// Flying, and Hungry Ghoul's sacrifice outlet.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9ccec56-ec2b-4ab8-b929-6e4795711db9",
		Name:            "Ravenous Harpy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.PutACounterOnThis(game.CounterPlusOne),
		}},
	})
}
