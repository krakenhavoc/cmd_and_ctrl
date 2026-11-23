package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seasoned Buttoneer — Creature — Vedalken Employee {2}{U}:
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
		OracleID:     "6048a682-f218-4039-8d5e-3e5156817fb2",
		Name:         "Seasoned Buttoneer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.WhenThisEntersOpenAnAttraction("Seasoned Buttoneer")},
	})
}
