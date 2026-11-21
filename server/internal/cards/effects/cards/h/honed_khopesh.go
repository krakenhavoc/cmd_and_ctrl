package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honed Khopesh — Artifact — Equipment {1}
// (Amonkhet, #5535):
//
//	"Equipped creature gets +1/+1.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)"
//
// The shared pump and equip constructors.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2142c41-ff66-4276-b426-fd945ed9b0f6",
		Name:         "Honed Khopesh",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
