package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystalline Crawler — Artifact Creature — Construct {4}, 1/1
// (Foundations Commander):
//
//	"Converge — This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.
//	 Remove a +1/+1 counter from this creature: Add one mana of any
//	 color.
//	 {T}: Put a +1/+1 counter on this creature."
//
// Rancorous Archaic's converge entry (CountersPerColorSpent, with the
// auto-tapper steered to distinct colours), Mana Bloom's counter-removal
// mana ability and Drake Hatcher's counter on itself. The mana ability
// has no {T}, so it works while the Crawler is summoning sick, and
// removing the counter is the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "5ea3e672-ae14-49a4-badb-f713497dd126",
		Name:                "Crystalline Crawler",
		Completeness:        CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersPerColorSpent(game.CounterPlusOne),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{RemoveCounters: RemoveCountersFromThis(game.CounterPlusOne, 1).RemoveCounters},
			Produced: "{W|U|B|R|G}",
			Label:    "Remove a +1/+1 counter from this creature: Add one mana of any color.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
