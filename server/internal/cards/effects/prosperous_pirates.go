package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prosperous Pirates — Creature — Human Pirate {4}{U}, 3/4 (Ixalan, #5532):
//
//	"When this creature enters, create two Treasure tokens."
//
// Plundering Pirate's body: the enters trigger uses the stack (CR 603.6a)
// and makes two Treasure tokens under its controller's control (CR 111.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9dc6281b-08e0-4c66-9aef-abc7184ca36a",
		Name:         "Prosperous Pirates",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Prosperous Pirates — create two Treasure tokens",
				Do(CreateToken{Template: TreasureToken(), N: 2})),
		},
	})
}
