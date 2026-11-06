package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gallant Pie-Wielder — Creature — Dwarf Knight {2}{W}, 2/3 (Wilds of
// Eldraine, #4471):
//
//	"First strike
//	 Celebration — This creature has double strike as long as two or more
//	 nonland permanents entered the battlefield under your control this
//	 turn."
//
// The printed first strike, plus a layer 6 double strike grant that is on
// while its controller has celebration (CR 207.2c, celebration.go): the
// tally is per turn, so it switches off at the next turn, and a
// permanent that entered and has since left still counts. Double strike
// beside first strike is the engine's combat damage (CR 702.4b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fba2fa87-3f2a-4a9f-a19e-a1f9dada4118",
		Name:            "Gallant Pie-Wielder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static:          []game.StaticAbility{SelfKeywordWhileCelebration("double strike")},
	})
}
