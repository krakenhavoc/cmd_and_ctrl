package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Concession Stand — Artifact — Attraction:
//
//	"Visit — Create a Food token. (It's an artifact with "{2}, {T}, Sacrifice
//	 this artifact: You gain 3 life.")"
//
// A visit ability (CR 702.159a): it triggers when its controller rolls to
// visit their Attractions and the result is lit up on this printing. The
// lights are the printing's (Card.AttractionLights, ADR 0207 §3), so two
// printings of Concession Stand can be visited on different numbers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2559fe2d-92b9-43c8-b538-d424df25eae8",
		Name:         "Concession Stand",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Visit("Concession Stand", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}
