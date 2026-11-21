package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Scimitar — Artifact — Equipment {1} (Foundations Jumpstart):
//
//	"Equipped creature gets +1/+1.
//	 Equip {1}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cde26d69-f3e7-4dd0-a53b-cd0ec812d717",
		Name:         "Leonin Scimitar",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
