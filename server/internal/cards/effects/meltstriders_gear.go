package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1d6742a3-592b-4e67-8cc6-56e2962ad7c7",
		Name:         "Meltstrider's Gear",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
			GrantToAttached("reach"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Meltstrider's Gear — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{5}")},
	})
}
