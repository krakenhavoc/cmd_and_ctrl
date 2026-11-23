package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chronomaton — Artifact Creature — Golem {1}, 1/1 (Magic 2013, #5568):
//
//	"{1}, {T}: Put a +1/+1 counter on this creature."
//
// Hangarback Walker's tap-to-grow row: the counter goes on the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "610f25b0-770f-457a-b0eb-6deb832f71d4",
		Name:         "Chronomaton",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  effects.PlusOneCountersOnThis(1),
		}},
	})
}
