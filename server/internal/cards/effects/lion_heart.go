package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lion Heart — Artifact — Equipment {4} (Final Fantasy, #4462):
//
//	"When this Equipment enters, it deals 2 damage to any target.
//	 Equipped creature gets +2/+1.
//	 Equip {2}"
//
// Meteorite's enters trigger (a target that left takes nothing, CR
// 608.2b) with Monk's Fist's pump (a layer 7c +2/+1 on the equipped
// creature) and equip.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd4151cc-743a-40f0-9e51-1f67555d9ab3",
		Name:         "Lion Heart",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Lion Heart — 2 damage to any target", sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
