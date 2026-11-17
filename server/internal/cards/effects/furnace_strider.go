package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furnace Strider — Creature — Phyrexian Beast {4}{R}, 4/5 (Phyrexia: All
// Will Be One):
//
//	"This creature enters with two oil counters on it.
//	 Remove an oil counter from this creature: Target creature you control
//	 gains haste until end of turn."
//
// The target is a creature its controller controls, itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85488433-3c23-4a29-b063-53dba71f26c4",
		Name:         "Furnace Strider",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(oilCounter, 2, "Furnace Strider: enters with two oil counters on it")},
		Activated: []ActivatedAbility{{
			Label:   "Remove an oil counter from this creature: Target creature you control gains haste until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    RemoveCountersFromThis(oilCounter, 1),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect:  ebTargetCreatureGainsUntilEOT("haste", "Furnace Strider — haste until end of turn"),
		}},
	})
}
