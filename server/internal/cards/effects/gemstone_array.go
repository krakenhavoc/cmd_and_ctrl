package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gemstone Array — Artifact {4} (Fifth Dawn, #5604):
//
//	"{2}: Put a charge counter on this artifact.
//	 Remove a charge counter from this artifact: Add one mana of any
//	 color."
//
// Black Mana Battery's charge counter without the tap, and Pentad Prism's
// "remove a charge counter: add one mana of any color" mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91303465-4b98-4ecc-82f4-cfd2f9a0f8a7",
		Name:         "Gemstone Array",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: Put a charge counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}"),
			Effect:  putChargeCounterOnThis,
		}},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{RemoveCounters: RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters},
			Produced: "{W|U|B|R|G}",
			Label:    "Remove a charge counter from this artifact: Add one mana of any color",
		}},
	})
}
