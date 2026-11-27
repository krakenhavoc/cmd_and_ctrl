package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spike Rogue — Creature — Spike {1}{G}{G}, 0/0 (Exodus, S82 OLD batch 741, #5626):
//
//	"This creature enters with two +1/+1 counters on it.
//	 {2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter
//	 on target creature.
//	 {2}, Remove a +1/+1 counter from a creature you control: Put a +1/+1
//	 counter on this creature."
//
// The cycle's entry replacement and counter move (spike_counters.go). The
// second ability takes its counter off a creature chosen as the cost is paid
// (Shapers of Nature): the Rogue or another. The counter goes on the Rogue
// only if it is still the same permanent as the ability resolves (CR 400.7).
//
// No simplification.
func init() {
	const name = "Spike Rogue"
	effects.Register(effects.Spec{
		OracleID:     "90ec956f-b6de-4214-9b5a-895280e542c1",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SpikeEntersWith(name, "two", 2)},
		Activated: []effects.ActivatedAbility{
			effects.SpikeMovesACounter(),
			{
				Label:   "{2}, Remove a +1/+1 counter from a creature you control: Put a +1/+1 counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.RemoveCountersFrom(game.CounterPlusOne, 1, "a creature you control", effects.Creature())),
				Effect:  effects.ThisStillHere(effects.PlusOneCountersOnThis(1)),
			},
		},
	})
}
