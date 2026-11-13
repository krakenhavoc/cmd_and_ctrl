package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Well of Discovery — Artifact {6} (Prophecy, #5794):
//
//	"At the beginning of your end step, if you control no untapped lands,
//	 draw a card."
//
// Well of Life's shape with a draw: an intervening "if" (CR 603.4,
// no_untapped_lands.go), checked as its controller's end step begins and
// again as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d63185ac-c7cd-4501-9248-54d6caf7d091",
		Name:         "Well of Discovery",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfNoUntappedLands("Well of Discovery — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
