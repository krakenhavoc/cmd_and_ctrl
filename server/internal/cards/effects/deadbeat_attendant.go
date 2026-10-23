package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadbeat Attendant — Creature — Vampire Employee {1}{B}:
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
		OracleID:     "ae4b00ab-25ad-4177-8b86-af8b84d67ef3",
		Name:         "Deadbeat Attendant",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{WhenThisEntersOpenAnAttraction("Deadbeat Attendant")},
	})
}
