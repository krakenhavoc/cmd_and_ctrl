package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulshok Battlegear — Artifact — Equipment {3} (Mirrodin, #5606):
//
//	"Equipped creature gets +3/+3.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// Meteor Sword's pump without the enters trigger: a layer 7c boost on
// the equipped creature (PumpAttached), and the shared sorcery-speed
// equip row (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "708df587-3b13-42cb-8341-f476ab4cbe45",
		Name:         "Vulshok Battlegear",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
