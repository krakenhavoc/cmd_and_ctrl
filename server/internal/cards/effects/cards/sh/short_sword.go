package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Short Sword — Artifact — Equipment {1} (Dominaria, #5530):
//
//	"Equipped creature gets +1/+1.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)"
//
// Leonin Scimitar's shape: +1/+1 and equip {1}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e842a4e5-ea0b-43ae-9ed8-93e30f4e7b8a",
		Name:         "Short Sword",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
