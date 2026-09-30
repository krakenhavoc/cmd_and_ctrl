package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Apex Devastator — Creature — Chimera Hydra {8}{G}{G}, 10/10:
//
//	"Cascade, cascade, cascade, cascade (When you cast this spell,
//	 exile cards from the top of your library until you exile a
//	 nonland card that costs less. You may cast it without paying its
//	 mana cost. Put the exiled cards on the bottom in a random order.
//	 Multiple instances of cascade each trigger separately.)"
//
// Four separate triggered abilities, not one that runs four times —
// Maelstrom Wanderer's own "cascade, cascade" doubled again. Each one
// goes on the stack independently when the Hydra is cast, the
// controller orders the four (CR 603.3b), and each digs on its own
// until it finds something costing less than ten.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4d6747c-3516-4e71-be30-af007c6b1dc4",
		Name:         "Apex Devastator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Cascade(), Cascade(), Cascade(), Cascade(),
		},
	})
}
