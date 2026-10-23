package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roller Coaster — Artifact — Attraction:
//
//	"Visit — Creatures you control get +2/+0 until end of turn."
//
// A visit ability (CR 702.159a): it triggers when its controller rolls to
// visit their Attractions and the result is lit up on this printing. The
// lights are the printing's (Card.AttractionLights, ADR 0207 §3), so two
// printings of Roller Coaster can be visited on different numbers.
// The pump fixes its set as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6efcf15a-2204-486f-9cff-7311054dcee0",
		Name:         "Roller Coaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Visit("Roller Coaster", Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Label: "Roller Coaster — creatures you control get +2/+0"})),
		},
	})
}
