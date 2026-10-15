package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dwarven Mauler — Creature — Dwarf Warrior {R}, 2/1 (#3778):
//
//	"Equip abilities you activate that target this creature cost {2}
//	 less to activate."
//
// Helitrooper's discount, alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70538676-2ffb-4cad-827e-b55e81088678",
		Name:         "Dwarven Mauler",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			EquipAbilitiesThatTargetThisCostLess(2,
				"Equip abilities you activate that target this creature cost {2} less to activate."),
		},
	})
}
