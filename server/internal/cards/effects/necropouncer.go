package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9e33b295-fab8-42c7-be43-baf9c17d063f",
		Name:         "Necropouncer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 1), GrantToAttached("haste")},
		Triggered:    []game.TriggeredAbility{LivingWeapon("Necropouncer")},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
