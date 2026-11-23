package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cf2cc44e-8134-4b37-bc10-694abb8808f0",
		Name:         "Spinny Ride",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Visit("Spinny Ride", effects.B36TapChosenCreature), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
