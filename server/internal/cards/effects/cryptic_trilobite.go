package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cryptic Trilobite — Creature — Trilobite {X}{X}, 0/0 (Commander 2020,
// #5513):
//
//	"This creature enters with X +1/+1 counters on it.
//	 Remove a +1/+1 counter from this creature: Add {C}{C}. Spend this mana
//	 only to activate abilities.
//	 {1}, {T}: Put a +1/+1 counter on this creature."
//
// The entry counters are Benevolent Hydra's XCounters off the cast's X. The
// mana ability is Crystalline Crawler's counter-removal cost with
// Redshift's "only to activate abilities" spend restriction (CR 106.6), so
// the {C}{C} pays for an activated ability of any source and never for a
// spell. The tap ability is Crystalline Crawler's with {1} added.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "28c0fd21-a88a-4a40-853f-2c1810517bc3",
		Name:                       "Cryptic Trilobite",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{RemoveCounters: RemoveCountersFromThis(game.CounterPlusOne, 1).RemoveCounters},
			Produced:     "{C}{C}",
			Label:        "Remove a +1/+1 counter from this creature: Add {C}{C}. Spend this mana only to activate abilities.",
			Restrictions: []string{game.ManaRestrictActivate},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
