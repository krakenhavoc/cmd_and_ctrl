package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "181965fb-e685-4b1f-9bd6-3a33bed933b2",
		Name:         "Meteor Sword",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Meteor Sword — destroy target permanent", destroyFirstLegalCardTarget),
				TargetPermanent("target permanent")),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
