package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aven Riftwatcher — Creature — Bird Rebel Soldier {2}{W}, 2/3 (Planar Chaos, #5595):
//
//	"Flying
//	 Vanishing 3 (This creature enters with three time counters on it. At the
//	 beginning of your upkeep, remove a time counter from it. When the last is
//	 removed, sacrifice it.)
//	 When this creature enters or leaves the battlefield, you gain 2 life."
//
// Vanishing is cmdr2B17Vanishing* (Deep Forest Hermit): the entry counters
// are a replacement, the upkeep removal and the last-counter sacrifice are
// two triggers (CR 702.63a).
// "Enters or leaves" is one ability with two trigger conditions
// (WhenThisEntersOrLeaves), so the vanishing sacrifice gains the second 2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a4d22abe-87c5-4f6b-ad14-4a7935883169",
		Name:            "Aven Riftwatcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{cmdr2B17VanishingEntry(3)},
		Triggered: append(cmdr2B17VanishingTriggers(),
			WhenThisEntersOrLeaves("Aven Riftwatcher — you gain 2 life", Do(GainLife{Amount: 2}))),
	})
}
