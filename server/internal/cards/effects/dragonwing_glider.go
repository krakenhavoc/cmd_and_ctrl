package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonwing Glider — Artifact — Equipment {3}{R}{R} (Phyrexia: All Will Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+2 and has flying and haste.
//	 Equip {3}{R}{R}"
//
// #5654, S82. The shared ForMirrodin trigger, a layer 7c +2/+2, a layer 6
// grant of both keywords and one equip ability. The Rebel has haste as
// soon as the Glider is on it, so it can attack the turn it is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4cd4c6da-a84e-4cba-893b-799a7cb5c411",
		Name:         "Dragonwing Glider",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{ForMirrodin("Dragonwing Glider")},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying", "haste"),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}{R}{R}")},
	})
}
