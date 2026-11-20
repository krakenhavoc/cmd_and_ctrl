package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burning-Tree Emissary — Creature — Human Shaman {R/G}{R/G}, 2/2 (Gatecrash, #5565):
//
//	"When this creature enters, add {R}{G}."
//
// Priest of Gix's shape: an enters trigger that adds mana as it
// resolves. It is not a mana ability (CR 605.1b), so it uses the stack;
// the mana empties at the end of the step (CR 500.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "327d9679-0049-4401-8dab-e0fb362306bd",
		Name:         "Burning-Tree Emissary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Burning-Tree Emissary — add {R}{G}", Do(AddMana{Produced: "{R}{G}"})),
		},
	})
}
