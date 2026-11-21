package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sailor of Means — Creature — Human Pirate {2}{U}, 1/4 (Ixalan, #5532):
//
//	"When this creature enters, create a Treasure token."
//
// Plundering Pirate's body: the enters trigger uses the stack (CR 603.6a)
// and makes a Treasure token under its controller's control (CR 111.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4efae03-99de-4439-925b-504602623bfd",
		Name:         "Sailor of Means",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sailor of Means — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
