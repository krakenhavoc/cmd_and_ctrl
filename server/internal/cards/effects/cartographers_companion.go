package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cartographer's Companion — Artifact Creature — Gnome {3}, 2/1 (The
// Lost Caverns of Ixalan, #4470):
//
//	"When this creature enters, create a Map token."
//
// The Map is the catalog's token (CR 111.10s), the same one Restless
// Anchorage and Get Lost make, with its own sorcery-speed explore
// ability. The enters trigger uses the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "460c0f0b-bbf4-4307-9b44-350175d83bfe",
		Name:         "Cartographer's Companion",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cartographer's Companion — create a Map token",
				Do(CreateToken{Template: MapToken(), N: 1})),
		},
	})
}
