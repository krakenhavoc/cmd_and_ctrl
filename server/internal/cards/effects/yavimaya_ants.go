package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yavimaya Ants — Creature — Insect {2}{G}{G}, 5/1 (Alliances, #5634):
//
//	"Trample, haste
//	 Cumulative upkeep {G}{G} (At the beginning of your upkeep, put an age
//	 counter on this permanent, then sacrifice it unless you pay its upkeep
//	 cost for each age counter on it.)"
//
// Trample and haste are engine keywords carried as printed data (CR
// 702.19b, CR 702.10b). Cumulative upkeep is the CumulativeUpkeep trigger
// (cumulative_upkeep.go): the age counter goes on first, then {G}{G} is
// charged once per counter, and an unpaid upkeep sacrifices the Ants
// (CR 702.24a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3f5bbfaa-cbd7-47af-9aa2-abb49d672734",
		Name:            "Yavimaya Ants",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Yavimaya Ants — cumulative upkeep {G}{G}", "{G}{G}"),
		},
	})
}
