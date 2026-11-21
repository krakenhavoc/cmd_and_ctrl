package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2f6c21a1-d618-497c-bfde-c0f2e1014d99",
		Name:         "Sylvok Battle-Chair",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ForMirrodin("Sylvok Battle-Chair")},
		Static: []game.StaticAbility{
			effects.PumpAttached(4, 4),
			effects.GrantToAttached("trample"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{5}{G}{G}")},
	})
}
