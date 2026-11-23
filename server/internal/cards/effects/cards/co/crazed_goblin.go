package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fafe2d90-e5fc-401a-a971-829192bec6c0",
		Name:         "Crazed Goblin",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
