package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1244902e-1a4e-4d3c-9a55-9be4602e3d2d",
		Name:         "Colossal Dreadmask",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Colossal Dreadmask"),
		},
		Static: []game.StaticAbility{
			PumpAttached(6, 6),
			GrantToAttached("trample"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}{G}{G}"),
		},
	})
}
