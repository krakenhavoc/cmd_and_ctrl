package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meteor Sword — Artifact — Equipment {7} (Avatar: The Last Airbender,
// #4459):
//
//	"When this Equipment enters, destroy target permanent.
//	 Equipped creature gets +3/+3.
//	 Equip {3}"
//
// The enter trigger is Bone Shredder's destroy with any permanent as the
// target, checked again on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "181965fb-e685-4b1f-9bd6-3a33bed933b2",
		Name:         "Meteor Sword",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Meteor Sword — destroy target permanent", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target permanent")),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
