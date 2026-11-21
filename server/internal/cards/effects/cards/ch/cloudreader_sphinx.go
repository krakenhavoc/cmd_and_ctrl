package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloudreader Sphinx — Creature — Sphinx {4}{U}, 3/4 (Dominaria, #5530):
//
//	"Flying
//	 When this creature enters, scry 2. (Look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)"
//
// Willow-Wind's shape: flying, and a scry 2 when it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b5a4bf91-4d31-4557-9ec4-a96ff4d9be6b",
		Name:            "Cloudreader Sphinx",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Scry: 2},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cloudreader Sphinx — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
