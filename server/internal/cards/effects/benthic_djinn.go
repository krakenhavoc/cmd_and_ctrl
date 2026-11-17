package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benthic Djinn — Creature — Djinn {2}{U}{B}, 5/3 (Mirage, #5633):
//
//	"Islandwalk (This creature can't be blocked as long as defending
//	 player controls an Island.)
//	 At the beginning of your upkeep, you lose 2 life."
//
// Islandwalk is a printed keyword the block legality check reads
// (CR 702.14c). The upkeep trigger is a life loss, not damage
// (CR 119.3), so no prevention effect stops it; it happens even if the
// Djinn has left the battlefield in response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7e8d40fd-c324-4ec2-897d-b46368b050e6",
		Name:            "Benthic Djinn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Benthic Djinn — you lose 2 life", Do(GainLife{Amount: -2})),
		},
	})
}
