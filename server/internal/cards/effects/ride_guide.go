package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ride Guide — Creature — Human Employee {4}{W}:
//
//	"When this creature enters, open an Attraction."
//
// Opening an Attraction is the engine's (ADR 0207,
// game.OpenAttractionsForEffect): the top card of its controller's
// Attraction deck enters under their control. A player playing without an
// Attraction deck, or with an empty one, opens nothing (CR 701.51a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8772e5cc-d3f3-4d68-86e7-6c105f39930a",
		Name:         "Ride Guide",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{WhenThisEntersOpenAnAttraction("Ride Guide")},
	})
}
