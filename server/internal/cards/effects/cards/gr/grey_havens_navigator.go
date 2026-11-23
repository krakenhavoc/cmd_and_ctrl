package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grey Havens Navigator — Creature — Elf Pilot {2}{U}, 3/2 (The Hobbit):
//
//	"Flash
//	 When this creature enters, scry 1."
//
// Flash lets it be cast at instant speed; the enters trigger scries its
// controller (CR 701.22), who may put the top card on the bottom.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "14291205-1ac2-4804-b365-4f65377d5803",
		Name:            "Grey Havens Navigator",
		Purpose:         game.Purpose{Scry: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Grey Havens Navigator — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
