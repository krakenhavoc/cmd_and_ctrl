package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "16e555f2-5aa8-4100-a036-eed48db0e84a",
		Name:         "Bone Saw",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0)},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}
