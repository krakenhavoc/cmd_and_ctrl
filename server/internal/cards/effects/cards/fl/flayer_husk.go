package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5cfdfcac-57fb-4507-a97a-cccbac2d1327",
		Name:         "Flayer Husk",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.LivingWeapon("Flayer Husk")},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
