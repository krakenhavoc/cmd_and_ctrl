package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran's Sidearm — Artifact — Equipment {2} (Magic Origins, #5549):
//
//	"Equipped creature gets +1/+1.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)"
//
// Leonin Scimitar's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f038c11e-b070-40fd-9d0f-e031f7bc2bd2",
		Name:         "Veteran's Sidearm",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
