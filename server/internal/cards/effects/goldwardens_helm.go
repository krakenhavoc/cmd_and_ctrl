package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goldwarden's Helm — Artifact — Equipment {2}{W} (Phyrexia: All Will Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equipped creature gets +0/+1.
//	 Equip {1}{W}"
//
// #5654, S82. The shared ForMirrodin trigger, a layer 7c +0/+1 and one
// equip ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c35f41a-55e4-49ca-b211-f57d7b1b90f6",
		Name:         "Goldwarden's Helm",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{ForMirrodin("Goldwarden's Helm")},
		Static: []game.StaticAbility{
			PumpAttached(0, 1),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}{W}")},
	})
}
