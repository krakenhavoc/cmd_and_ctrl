package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fortune Teller — Artifact — Attraction:
//
//	"Visit — Scry 1."
//
// A visit ability (CR 702.159a): it triggers when its controller rolls to
// visit their Attractions and the result is lit up on this printing. The
// lights are the printing's (Card.AttractionLights, ADR 0207 §3), so two
// printings of Fortune Teller can be visited on different numbers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a837fe34-ae85-4bb6-af5a-4dec30b18f52",
		Name:         "Fortune Teller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Visit("Fortune Teller", effects.Do(effects.Scry{N: 1})),
		},
	})
}
