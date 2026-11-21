package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulshok Splitter — Artifact — Equipment {3}{R} (Phyrexia: All Will Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+0.
//	 Equip {2}{R}"
//
// #5654, S82. The shared ForMirrodin trigger, a layer 7c +2/+0 and one
// equip ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6253f9d4-83ff-4a0b-b465-e9f9a565f664",
		Name:         "Vulshok Splitter",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ForMirrodin("Vulshok Splitter")},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{R}")},
	})
}
