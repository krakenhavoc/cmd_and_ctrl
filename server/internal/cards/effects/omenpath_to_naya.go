package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omenpath to Naya — Land — Omenpath:
//
//	"Vanishing 4 (This land enters the battlefield with four time counters
//	 on it. At the beginning of your upkeep, remove a time counter from it.
//	 When the last is removed, sacrifice it.)
//	 {T}: Add {R}, {G}, or {W}."
//
// Vanishing is the shared three abilities (cmdr2B17Vanishing*, CR 702.63a):
// the entry counters, the upkeep removal and the last-counter sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "7a82c183-1561-4276-9a75-6bbb9c2bee2e",
		Name:          "Omenpath to Naya",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{cmdr2B17VanishingEntry(4)},
		Triggered:     cmdr2B17VanishingTriggers(),
		ManaAbilities: []ManaAbility{b03TriMana("R", "G", "W")},
	})
}
