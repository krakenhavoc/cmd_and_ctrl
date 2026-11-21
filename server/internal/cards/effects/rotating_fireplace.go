package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rotating Fireplace — Artifact {3} (Doctor Who, #5948):
//
//	"This artifact enters tapped with a time counter on it.
//	 {T}: Add an amount of {C} equal to the number of time counters on this
//	 artifact.
//	 {4}, {T}: Time travel. Activate only as a sorcery."
//
// The entry replacement is the shared enters-tapped-with-counters row. The
// mana ability counts the time counters as it resolves
// (ProducedPerCounterOnThis); time travel is the shared primitive (CR 701.56a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f07a8c8d-73a1-44cc-9313-05f3bcf1a67c",
		Name:         "Rotating Fireplace",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedWithCounters(game.CounterTime, 1)},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedPerCounterOnThis(game.CounterTime, "C"),
			Label:          "Add {C} for each time counter on this artifact",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{4}, {T}: Time travel. Activate only as a sorcery.",
			Cost:         game.AbilityCost{Mana: "{4}", Tap: true},
			SorcerySpeed: true,
			Effect:       Do(TimeTravel{}),
		}},
	})
}
