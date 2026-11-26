package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malevolent Noble — Creature — Human Noble {1}{B}, 2/2 (Throne of
// Eldraine, #5516):
//
//	"{2}, Sacrifice an artifact or another creature: Put a +1/+1 counter
//	 on this creature."
//
// The sacrifice is a cost (CR 602.2b); "another" is object identity, so the
// Noble itself is not a legal choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c23ae3a-108f-4a09-b89f-8c43c33dd144",
		Name:         "Malevolent Noble",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice an artifact or another creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeAnotherN(1, "an artifact or another creature", effects.Or(effects.Artifact(), effects.Creature()))),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
