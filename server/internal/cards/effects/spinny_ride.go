package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spinny Ride — Artifact — Attraction:
//
//	"Visit — Tap target creature an opponent controls."
//
// A visit ability (CR 702.159a): it triggers when its controller rolls to
// visit their Attractions and the result is lit up on this printing. The
// lights are the printing's (Card.AttractionLights, ADR 0207 §3), so two
// printings of Spinny Ride can be visited on different numbers.
// Targeted, so with no creature an opponent controls the visit is
// removed from the stack as it would be put there (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf2cc44e-8134-4b37-bc10-694abb8808f0",
		Name:         "Spinny Ride",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Visit("Spinny Ride", b36TapChosenCreature), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
