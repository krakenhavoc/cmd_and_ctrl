package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meltstrider's Gear — Artifact — Equipment {G} (Edge of Eternities,
// #4461):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +2/+1 and has reach.
//	 Equip {5} ({5}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// Falcon's Wing Harness's shape: the entry trigger attaches it for free
// (AttachSourceToTarget), PumpAttached and GrantToAttached are the
// equipped creature's bonus, and EquipAbility is the CR 702.6a equip,
// sorcery speed. With no creature of yours to target the trigger is
// removed and the Gear waits for its equip.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d6742a3-592b-4e67-8cc6-56e2962ad7c7",
		Name:         "Meltstrider's Gear",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 1),
			effects.GrantToAttached("reach"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Meltstrider's Gear — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{5}")},
	})
}
