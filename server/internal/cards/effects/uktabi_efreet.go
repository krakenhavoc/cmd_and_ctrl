package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Uktabi Efreet — Creature — Efreet {2}{G}{G}, 5/4 (Weatherlight, #5629):
//
//	"Cumulative upkeep {G} (At the beginning of your upkeep, put an age
//	 counter on this permanent, then sacrifice it unless you pay its upkeep
//	 cost for each age counter on it.)"
//
// The shared CumulativeUpkeep trigger (CR 702.24a): each of your
// upkeeps an age counter goes on, then {G} is asked once per counter,
// and declining sacrifices it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fd9a37f-151a-4501-9d8e-bc67cc902592",
		Name:         "Uktabi Efreet",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Uktabi Efreet — cumulative upkeep {G}", "{G}"),
		},
	})
}
