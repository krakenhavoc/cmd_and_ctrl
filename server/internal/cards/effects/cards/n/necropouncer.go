package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necropouncer — Artifact — Equipment {6} (New Phyrexia, #5575):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +3/+1 and has haste.
//	 Equip {2}"
//
// Living weapon (CR 702.92) and Boots of Speed's statics at +3/+1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e33b295-fab8-42c7-be43-baf9c17d063f",
		Name:         "Necropouncer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 1), effects.GrantToAttached("haste")},
		Triggered:    []game.TriggeredAbility{effects.LivingWeapon("Necropouncer")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
