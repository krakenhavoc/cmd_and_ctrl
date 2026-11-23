package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waning Wurm — Creature — Zombie Wurm {3}{B}, 7/6 (Planar Chaos, #5595):
//
//	"Vanishing 2 (This creature enters with two time counters on it. At the
//	 beginning of your upkeep, remove a time counter from it. When the last is
//	 removed, sacrifice it.)"
//
// Vanishing is cmdr2B17Vanishing* (Deep Forest Hermit): the entry counters
// are a replacement, the upkeep removal and the last-counter sacrifice are
// two triggers (CR 702.63a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4921a314-fce7-4f3c-8546-6cf520a2c104",
		Name:         "Waning Wurm",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Cmdr2B17VanishingEntry(2)},
		Triggered:    effects.Cmdr2B17VanishingTriggers(),
	})
}
