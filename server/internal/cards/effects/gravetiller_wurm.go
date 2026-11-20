package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gravetiller Wurm — Creature — Wurm {5}{G}, 4/4 (Dark Ascension, #5571):
//
//	"Trample
//	 Morbid — This creature enters with four +1/+1 counters on it if
//	 a creature died this turn."
//
// Festerhide Boar's morbid replacement (CR 614.1c) with four counters,
// read as the Wurm enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "020d0c91-da68-40da-8561-5522377f3d07",
		Name:            "Gravetiller Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    []game.ReplacementEffect{SelfEntersWithCountersIfMorbid(game.CounterPlusOne, 4)},
	})
}
