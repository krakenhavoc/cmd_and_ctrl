package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bone Saw — Artifact — Equipment for {0}:
//
//	"Equipped creature gets +1/+0.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// The cheapest Equipment in the catalog to cast, and the simplest to
// implement — a plain layer 7c pump on the host and the ordinary
// equip ability. Nothing else printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16e555f2-5aa8-4100-a036-eed48db0e84a",
		Name:         "Bone Saw",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0)},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
