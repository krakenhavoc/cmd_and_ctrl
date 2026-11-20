package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Myr Moonvessel — Artifact Creature — Myr {1} (Darksteel, #5605):
//
//	"When this creature dies, add {C}."
//
// Cathodion's dies trigger with one {C}. It uses the stack (it is not a
// mana ability, CR 605.1b), and the mana empties at the end of the step
// like any other (CR 500.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e922661-80df-4e84-a12a-524bc74e6c9d",
		Name:         "Myr Moonvessel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Myr Moonvessel — add {C}", Do(AddMana{Produced: "{C}"})),
		},
	})
}
