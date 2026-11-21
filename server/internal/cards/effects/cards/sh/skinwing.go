package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skinwing — Artifact — Equipment {4} (Mirrodin Besieged,
// #5576):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +2/+2 and has flying.
//	 Equip {6}"
//
// Batterskull's living weapon and equip shape (LivingWeapon, CR 702.92a)
// with a +2/+2 pump and a layer-6 flying grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "100c877d-4cce-4b96-8f21-a4f8a43e7d84",
		Name:         "Skinwing",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.LivingWeapon("Skinwing")},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{6}")},
	})
}
