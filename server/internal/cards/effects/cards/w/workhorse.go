package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Workhorse — Artifact Creature — Horse {6}, 0/0 (Exodus, S82 OLD batch 741, #5626):
//
//	"This creature enters with four +1/+1 counters on it.
//	 Remove a +1/+1 counter from this creature: Add {C}."
//
// The Spike cycle's entry replacement (spike_counters.go) and a mana ability
// (CR 605.1a) whose only cost is the counter (CR 118.3): no {T}, so it works
// summoning-sick, once per counter. Removing the last one leaves a 0/0.
//
// No simplification.
func init() {
	const name = "Workhorse"
	effects.Register(effects.Spec{
		OracleID:     "bb4ab425-3da7-455d-b217-3d585d476113",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SpikeEntersWith(name, "four", 4)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{RemoveCounters: effects.RemoveCountersFromThis(game.CounterPlusOne, 1).RemoveCounters},
			Produced: "{C}",
			Label:    "Remove a +1/+1 counter from this creature: Add {C}.",
		}},
	})
}
