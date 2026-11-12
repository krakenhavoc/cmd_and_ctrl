package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirran Bardiche — Artifact — Equipment {4}{W} (Phyrexia: All Will Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+1 and has vigilance.
//	 Equip {3}{W}"
//
// #5654, S82. The shared ForMirrodin trigger, a layer 7c +2/+1, a layer 6
// vigilance grant and one equip ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1fd04d9e-cf89-49ba-a932-5d278f65bf77",
		Name:         "Mirran Bardiche",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{ForMirrodin("Mirran Bardiche")},
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
			GrantToAttached("vigilance"),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}{W}")},
	})
}
