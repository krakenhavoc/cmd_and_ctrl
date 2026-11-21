package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sickleslicer — Artifact — Equipment {3} (New Phyrexia, #5575):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +2/+2.
//	 Equip {4}"
//
// Living weapon (CR 702.92) and a plain +2/+2 Equipment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a004a4b-b53b-4b8e-86f5-528c1082e146",
		Name:         "Sickleslicer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered:    []game.TriggeredAbility{effects.LivingWeapon("Sickleslicer")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
