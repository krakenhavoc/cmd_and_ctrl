package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flayer Husk — Artifact — Equipment {1} (Mirrodin Besieged,
// #5576):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +1/+1.
//	 Equip {2}"
//
// Batterskull's living weapon and equip shape (LivingWeapon, CR 702.92a)
// with a +1/+1 pump: the Germ is a 1/1 while it is equipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5cfdfcac-57fb-4507-a97a-cccbac2d1327",
		Name:         "Flayer Husk",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{LivingWeapon("Flayer Husk")},
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
