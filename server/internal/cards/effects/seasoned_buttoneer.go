package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seasoned Buttoneer — Creature — Vedalken Employee {2}{U}:
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
		OracleID:     "6048a682-f218-4039-8d5e-3e5156817fb2",
		Name:         "Seasoned Buttoneer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{WhenThisEntersOpenAnAttraction("Seasoned Buttoneer")},
	})
}
