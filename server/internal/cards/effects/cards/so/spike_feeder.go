package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spike Feeder — Creature — Spike {1}{G}{G}, 0/0 (Stronghold, #5627):
//
//	"This creature enters with two +1/+1 counters on it.
//	 {2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter
//	 on target creature.
//	 Remove a +1/+1 counter from this creature: You gain 2 life."
//
// The cycle's shared entry replacement (CR 614.1c) and counter move
// (spike_counters.go). The life ability's only cost is the counter
// (CR 118.3), so it can be activated at instant speed, summoning sick or
// not, once per counter; removing the last one still gains the 2 life,
// though the Feeder dies as a 0/0 (CR 704.5f).
//
// No simplification.
func init() {
	const name = "Spike Feeder"
	effects.Register(effects.Spec{
		OracleID:     "5b147eab-39dd-4c2a-8f59-d778037788aa",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SpikeEntersWith(name, "two", 2)},
		Activated: []effects.ActivatedAbility{
			effects.SpikeMovesACounter(),
			{
				Label:   "Remove a +1/+1 counter from this creature: You gain 2 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.RemoveCountersFromThis(game.CounterPlusOne, 1),
				Effect:  effects.Do(effects.GainLife{Amount: 2}),
			},
		},
	})
}
