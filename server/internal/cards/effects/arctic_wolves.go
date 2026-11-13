package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arctic Wolves — Creature — Wolf {3}{G}{G}, 4/5 (Weatherlight, #5629):
//
//	"Cumulative upkeep {2} (At the beginning of your upkeep, put an age
//	 counter on this permanent, then sacrifice it unless you pay its upkeep
//	 cost for each age counter on it.)
//	 When this creature enters, draw a card."
//
// The shared CumulativeUpkeep trigger (CR 702.24a) and an enters
// trigger that draws a card when it resolves, as on Elvish Visionary.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9e7218f-07f2-46f3-bc63-eb5c0d32538b",
		Name:         "Arctic Wolves",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Arctic Wolves — cumulative upkeep {2}", "{2}"),
			WhenThisEnters("Arctic Wolves — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
