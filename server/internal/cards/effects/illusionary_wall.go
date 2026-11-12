package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illusionary Wall — Creature — Illusion Wall {4}{U}, 7/4 (Ice Age,
// #5636):
//
//	"Defender, flying, first strike
//	 Cumulative upkeep {U}"
//
// Defender, flying and first strike are printed keywords: it can't
// attack (CR 702.3b), and it blocks fliers and deals its damage first.
// Cumulative upkeep is the shared CumulativeUpkeep trigger
// (CR 702.24a): {U} per age counter each of your upkeeps, or it is
// sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "77a8945f-4d52-4445-8913-425007e7948c",
		Name:            "Illusionary Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Illusionary Wall — cumulative upkeep {U}", "{U}"),
		},
	})
}
