package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Predation Steward — Creature — Phyrexian Elf Warrior {1}{G}, 2/2
// (Phyrexia: All Will Be One):
//
//	"This creature enters with two oil counters on it.
//	 {2}{G}, {T}, Remove an oil counter from this creature: Target creature
//	 gets +2/+2 until end of turn. Activate only as a sorcery."
//
// Any creature, yours or not, itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e3d2b03-4c44-49a3-ab11-871ca9c372f7",
		Name:         "Predation Steward",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(oilCounter, 2, "Predation Steward: enters with two oil counters on it")},
		Activated: []ActivatedAbility{{
			Label:        "{2}{G}, {T}, Remove an oil counter from this creature: Target creature gets +2/+2 until end of turn. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
			Cost:         Plus(ManaCost("{2}{G}"), TapCost(), RemoveCountersFromThis(oilCounter, 1)),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       boostTargetUntilEOT("Predation Steward — +2/+2", 2, 2),
		}},
	})
}
