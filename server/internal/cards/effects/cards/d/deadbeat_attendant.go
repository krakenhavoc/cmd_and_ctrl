package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadbeat Attendant — Creature — Vampire Employee {1}{B}:
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
	effects.Register(effects.Spec{
		OracleID:     "ae4b00ab-25ad-4177-8b86-af8b84d67ef3",
		Name:         "Deadbeat Attendant",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.WhenThisEntersOpenAnAttraction("Deadbeat Attendant")},
	})
}
