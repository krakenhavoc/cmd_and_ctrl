package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Mauler — Creature — Dwarf Warrior {R}, 2/1 (#3778):
//
//	"Equip abilities you activate that target this creature cost {2}
//	 less to activate."
//
// Helitrooper's discount, alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70538676-2ffb-4cad-827e-b55e81088678",
		Name:         "Dwarven Mauler",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.EquipAbilitiesThatTargetThisCostLess(2,
				"Equip abilities you activate that target this creature cost {2} less to activate."),
		},
	})
}
