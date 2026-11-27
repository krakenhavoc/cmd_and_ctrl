package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gobbling Ooze — Creature — Ooze {4}{G}, 3/3 (Return to Ravnica, #5566):
//
//	"{G}, Sacrifice another creature: Put a +1/+1 counter on this creature."
//
// The sacrifice is a cost (CR 118.3) so it is paid before the ability goes on the
// stack and cannot be the Ooze itself. The counter goes on the Ooze only while it
// is still the permanent the ability came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "765a12fc-9b66-4bb7-88f6-6e7bd2d2f8c7",
		Name:         "Gobbling Ooze",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice another creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
