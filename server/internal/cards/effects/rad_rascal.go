package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rad Rascal — Creature — Devil Employee {3}{R}:
//
//	"When this creature enters, open an Attraction."
//
// Opening an Attraction is the engine's (ADR 0207,
// game.OpenAttractionsThenForEffect): the top card of its controller's
// Attraction deck enters under their control. A player playing without an
// Attraction deck, or with an empty one, opens nothing (CR 701.51a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66d4a218-5404-4389-b495-eaeda9a34401",
		Name:         "Rad Rascal",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{WhenThisEntersOpenAnAttraction("Rad Rascal")},
	})
}
