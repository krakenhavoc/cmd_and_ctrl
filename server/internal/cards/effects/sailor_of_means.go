package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sailor of Means — Creature — Human Pirate {2}{U}, 1/4 (Ixalan, #5532):
//
//	"When this creature enters, create a Treasure token."
//
// Plundering Pirate's body: the enters trigger uses the stack (CR 603.6a)
// and makes a Treasure token under its controller's control (CR 111.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4efae03-99de-4439-925b-504602623bfd",
		Name:         "Sailor of Means",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sailor of Means — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
