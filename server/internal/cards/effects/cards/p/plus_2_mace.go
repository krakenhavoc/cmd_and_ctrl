package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// +2 Mace — Artifact — Equipment {1}{W} (Adventures in the Forgotten
// Realms, #5499):
//
//	Equipped creature gets +2/+2.
//	Equip {3}
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "629fe1be-272d-465f-b9b1-2ce177410f13",
		Name:         "+2 Mace",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
