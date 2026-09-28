package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nemesis Mask — Artifact — Equipment, {3}:
//
//	"All creatures able to block equipped creature do so.
//	 Equip {3} ({3}: Attach to target creature you control. Equip
//	 only as a sorcery.)"
//
// #1684 leftover: the Lure clause on an Equipment is exactly Lure's
// own shape — BlockRequirementWhere(game.BlockRequirementLure,
// AttachedToSource) — plus an ordinary EquipAbility. No new
// machinery.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34062d79-457f-4af3-b9c6-cab2e4d17581",
		Name:         "Nemesis Mask",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			BlockRequirementWhere(game.BlockRequirementLure, AttachedToSource),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
