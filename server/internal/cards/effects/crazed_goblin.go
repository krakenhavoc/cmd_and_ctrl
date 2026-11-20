package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crazed Goblin — Creature — Goblin Warrior {R} (Darksteel, #5605):
//
//	"This creature attacks each combat if able."
//
// Bloodrock Cyclops's requirement (CR 508.1d): the shared
// AttacksEachCombat static. A tapped or summoning-sick Goblin, or one a
// cost stands in front of, is not "able", and is not forced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fafe2d90-e5fc-401a-a971-829192bec6c0",
		Name:         "Crazed Goblin",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
	})
}
