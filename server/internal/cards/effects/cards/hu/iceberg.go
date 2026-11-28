package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iceberg — Enchantment {X}{U}{U} (Ice Age, S82 OLD batch 845, #5636):
//
//	"This enchantment enters with X ice counters on it.
//	 {3}: Put an ice counter on this enchantment.
//	 Remove an ice counter from this enchantment: Add {C}."
//
// The X entry counters are the CR 614.1c clause (XCounters, Mikaeus's
// shape); the mana ability's only cost is the counter (Workhorse), so it
// has no {T} and works the turn it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "cff3fba2-9c5d-4e98-9ff0-0cc0b3bc8dbe",
		Name:                       "Iceberg",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters("ice")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Put an ice counter on this enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}"),
			Effect:  effects.PutACounterOnThis("ice"),
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{RemoveCounters: effects.RemoveCountersFromThis("ice", 1).RemoveCounters},
			Produced: "{C}",
			Label:    "Remove an ice counter from this enchantment: Add {C}.",
		}},
	})
}
