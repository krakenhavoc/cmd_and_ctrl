package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illusionary Forces — Creature — Illusion {3}{U}, 4/4 (Ice Age, #5636):
//
//	"Flying
//	 Cumulative upkeep {U}"
//
// Flying is a printed keyword. Cumulative upkeep is the shared
// CumulativeUpkeep trigger (CR 702.24a): each of your upkeeps an age
// counter goes on, then {U} is asked once per counter, and declining
// sacrifices it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e17e3bda-75cb-47a1-9f60-f50705a46a1e",
		Name:            "Illusionary Forces",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Illusionary Forces — cumulative upkeep {U}", "{U}"),
		},
	})
}
