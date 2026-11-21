package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "1fd04d9e-cf89-49ba-a932-5d278f65bf77",
		Name:         "Mirran Bardiche",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ForMirrodin("Mirran Bardiche")},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 1),
			effects.GrantToAttached("vigilance"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}{W}")},
	})
}
