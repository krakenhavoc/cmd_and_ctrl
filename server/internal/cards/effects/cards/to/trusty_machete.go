package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trusty Machete — Artifact — Equipment {1} (Zendikar, #5583):
//
//	"Equipped creature gets +2/+1.
//	 Equip {2}"
//
// The shared attachment pump and the sorcery-speed EquipAbility.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "394fae8f-3757-4e08-b97c-5d7c451f4e72",
		Name:         "Trusty Machete",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
