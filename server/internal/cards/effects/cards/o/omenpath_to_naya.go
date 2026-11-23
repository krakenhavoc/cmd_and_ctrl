package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "7a82c183-1561-4276-9a75-6bbb9c2bee2e",
		Name:          "Omenpath to Naya",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.Cmdr2B17VanishingEntry(4)},
		Triggered:     effects.Cmdr2B17VanishingTriggers(),
		ManaAbilities: []effects.ManaAbility{effects.B03TriMana("R", "G", "W")},
	})
}
