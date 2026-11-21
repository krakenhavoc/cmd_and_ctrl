package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbed Batterfist — Artifact — Equipment {1}{R} (Phyrexia: All Will Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/-1.
//	 Equip {1}"
//
// #5654, S82. The shared ForMirrodin trigger, a layer 7c +1/-1 and one
// equip ability. The Rebel it arrives on is a 3/1; moved onto a creature
// with 1 toughness it kills it (CR 704.5f), which is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93c2b91d-7d4e-483c-a11f-4a525f881503",
		Name:         "Barbed Batterfist",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ForMirrodin("Barbed Batterfist")},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, -1),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
