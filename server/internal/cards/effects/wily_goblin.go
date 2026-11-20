package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wily Goblin — Creature — Goblin Pirate {R}{R}, 1/1 (Ixalan, #5532):
//
//	"When this creature enters, create a Treasure token."
//
// Plundering Pirate's body: the enters trigger uses the stack (CR 603.6a)
// and makes a Treasure token under its controller's control (CR 111.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d27ffb4c-3a34-41aa-b464-6b897132c4ea",
		Name:         "Wily Goblin",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wily Goblin — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
