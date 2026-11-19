package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Festerhide Boar — Creature — Boar {3}{G}, 3/3 (Innistrad Remastered,
// S82 GAP, #5948):
//
//	"Trample
//	 Morbid — This creature enters with two +1/+1 counters on it if a
//	 creature died this turn."
//
// Morbid is an ability word (CR 207.2c) here on an entry replacement
// (CR 614.1c), Putrid Pals's shape with morbid's condition: any
// creature's death this turn, tokens included, read as the Boar enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be1fc2cf-00b8-424f-b683-3743fba285c9",
		Name:            "Festerhide Boar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    []game.ReplacementEffect{SelfEntersWithCountersIfMorbid(game.CounterPlusOne, 2)},
	})
}
