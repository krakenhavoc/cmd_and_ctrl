package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2155f3d8-c876-42bc-b546-a476425a4133",
		Name:         "Information Booth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Visit("Information Booth", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
