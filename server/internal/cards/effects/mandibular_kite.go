package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mandibular Kite — Artifact — Equipment {W} (Modern Horizons 3,
// #5467):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +1/+1 and has flying.
//	 Equip {3}{W}"
//
// Batterbone's shape: the shared LivingWeapon trigger (CR 702.92a), the
// layer-7c pump and the layer-6 grant every Equipment uses, and one
// equip ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bca9f65e-ea7a-4b8d-ab0f-c2672fa967f0",
		Name:         "Mandibular Kite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Mandibular Kite"),
		},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("flying"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}{W}"),
		},
	})
}
