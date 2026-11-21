package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Octoprophet — Creature — Octopus {3}{U}, 3/3 (Core Set 2020, #5518):
//
//	"When this creature enters, scry 2. (Look at the top two cards of your
//	 library, then put any number of them on the bottom and the rest on top
//	 in any order.)"
//
// Automatic Librarian's enters scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af6cfc23-b64b-4c8a-94b4-69d2b01cb3b9",
		Name:         "Octoprophet",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Octoprophet — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
