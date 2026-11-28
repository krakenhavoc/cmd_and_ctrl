package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spike Hatcher — Creature — Spike {6}{G}, 0/0 (Exodus, S82 OLD batch 741, #5626):
//
//	"This creature enters with six +1/+1 counters on it.
//	 {2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter
//	 on target creature.
//	 {1}, Remove a +1/+1 counter from this creature: Regenerate this
//	 creature."
//
// The Spike cycle's entry replacement and counter move (spike_counters.go);
// the regeneration is paid with a counter and {1} (CR 118.3), so it works at
// instant speed, and gives the shield (CR 701.19a) even if that was the last
// counter (the 0/0 then dies at once, CR 704.5f, shield or not).
//
// No simplification.
func init() {
	const name = "Spike Hatcher"
	effects.Register(effects.Spec{
		OracleID:     "726fe7a4-49b3-4b67-9d8c-6b334d0fd915",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SpikeEntersWith(name, "six", 6)},
		Activated: []effects.ActivatedAbility{
			effects.SpikeMovesACounter(),
			{
				Label:   "{1}, Remove a +1/+1 counter from this creature: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
				Effect:  effects.RegenerateThisSource,
			},
		},
	})
}
