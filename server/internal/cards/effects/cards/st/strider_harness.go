package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strider Harness — Artifact — Equipment {3} (Scars of Mirrodin, #5577):
//
//	"Equipped creature gets +1/+1 and has haste.
//	 Equip {1}"
//
// Boots of Speed's shape: PumpAttached, GrantToAttached and the equip
// row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd4ad383-8c40-4dda-bc9c-ba8a001d6882",
		Name:         "Strider Harness",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1), effects.GrantToAttached("haste")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
