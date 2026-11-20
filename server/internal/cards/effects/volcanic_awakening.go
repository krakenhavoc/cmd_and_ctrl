package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volcanic Awakening — Sorcery {4}{R}{R} (Time Spiral, #5596):
//
//	"Destroy target land.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. You may choose new targets for the copies.)"
//
// Stone Rain with the engine's storm trigger (CR 702.40a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2145329d-49df-4f4a-b7d7-3cc2f64639d3",
		Name:         "Volcanic Awakening",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve:    destroyTheTargetPermanent,
	})
}
