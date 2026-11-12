package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvok Battle-Chair — Artifact — Equipment {4}{G}{G} (Phyrexia: All Will Be
// One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equipped creature gets +4/+4 and has trample.
//	 Equip {5}{G}{G}"
//
// #5654, S82. The shared ForMirrodin trigger, a layer 7c +4/+4, a layer 6
// trample grant and one equip ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f6c21a1-d618-497c-bfde-c0f2e1014d99",
		Name:         "Sylvok Battle-Chair",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{ForMirrodin("Sylvok Battle-Chair")},
		Static: []game.StaticAbility{
			PumpAttached(4, 4),
			GrantToAttached("trample"),
		},
		Activated: []ActivatedAbility{EquipAbility("{5}{G}{G}")},
	})
}
