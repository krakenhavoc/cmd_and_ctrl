package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kiddie Coaster — Artifact — Attraction:
//
//	"Visit — Creatures you control get +1/+0 until end of turn."
//
// A visit ability (CR 702.159a): it triggers when its controller rolls to
// visit their Attractions and the result is lit up on this printing. The
// lights are the printing's (Card.AttractionLights, ADR 0207 §3), so two
// printings of Kiddie Coaster can be visited on different numbers.
// The pump fixes its set as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c99ce798-e332-4672-8fba-6da74ab14582",
		Name:         "Kiddie Coaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Visit("Kiddie Coaster", Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: "Kiddie Coaster — creatures you control get +1/+0"})),
		},
	})
}
