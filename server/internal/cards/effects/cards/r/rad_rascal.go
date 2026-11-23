package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rad Rascal — Creature — Devil Employee {3}{R}:
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
		OracleID:     "66d4a218-5404-4389-b495-eaeda9a34401",
		Name:         "Rad Rascal",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.WhenThisEntersOpenAnAttraction("Rad Rascal")},
	})
}
