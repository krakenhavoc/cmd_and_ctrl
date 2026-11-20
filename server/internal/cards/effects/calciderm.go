package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Calciderm — Creature — Beast {2}{W}{W}, 5/5 (Planar Chaos, #5595):
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Vanishing 4 (This creature enters with four time counters on it. At the
//	 beginning of your upkeep, remove a time counter from it. When the last is
//	 removed, sacrifice it.)"
//
// Shroud is an engine keyword carried as printed data (CR 702.18a).
// Vanishing is cmdr2B17Vanishing* (Deep Forest Hermit): the entry counters
// are a replacement, the upkeep removal and the last-counter sacrifice are
// two triggers (CR 702.63a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "329d53eb-d443-4376-ae5a-4be7dfb1fc91",
		Name:            "Calciderm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Replacements:    []game.ReplacementEffect{cmdr2B17VanishingEntry(4)},
		Triggered:       cmdr2B17VanishingTriggers(),
	})
}
