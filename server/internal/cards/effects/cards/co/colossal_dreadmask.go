package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Colossal Dreadmask — Artifact — Equipment {4}{G}{G} (Modern Horizons 3,
// #5467):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +6/+6 and has trample.
//	 Equip {3}{G}{G}"
//
// Batterbone's shape: the shared LivingWeapon trigger (CR 702.92a), the
// layer-7c pump and the layer-6 grant every Equipment uses, and one
// equip ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1244902e-1a4e-4d3c-9a55-9be4602e3d2d",
		Name:         "Colossal Dreadmask",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LivingWeapon("Colossal Dreadmask"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(6, 6),
			effects.GrantToAttached("trample"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{3}{G}{G}"),
		},
	})
}
