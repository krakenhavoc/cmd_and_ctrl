package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viridian Claw — Artifact — Equipment {2} (Mirrodin Besieged,
// #5576):
//
//	"Equipped creature gets +1/+0 and has first strike.
//	 Equip {1}"
//
// A layer 7c +1/+0 (PumpAttached) and a layer-6 first strike grant
// (GrantToAttached) on the equipped creature, with the equip ability
// (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11ba27f4-decf-4447-9808-0d58844655cf",
		Name:         "Viridian Claw",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("first strike"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
