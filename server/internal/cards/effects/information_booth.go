package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Information Booth — Artifact — Attraction:
//
//	"Visit — Draw a card."
//
// A visit ability (CR 702.159a): it triggers when its controller rolls to
// visit their Attractions and the result is lit up on this printing. The
// lights are the printing's (Card.AttractionLights, ADR 0207 §3), so two
// printings of Information Booth can be visited on different numbers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2155f3d8-c876-42bc-b546-a476425a4133",
		Name:         "Information Booth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Visit("Information Booth", Do(DrawCards{N: 1})),
		},
	})
}
