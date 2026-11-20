package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daru Warchief — Creature — Human Soldier {2}{W}{W}, 1/1 (Scourge):
//
//	"Soldier spells you cast cost {1} less to cast.
//	 Soldier creatures you control get +1/+2."
//
// Undead Warchief's discount for Soldiers (one cost modifier, CR 601.2f,
// generic mana only) and a layer-7c anthem (CR 613.4c) over Soldier
// creatures you control, the Warchief included, so it is a 2/3 on its own.
//
// No simplification.
func init() {
	anthem := effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Soldier"}, YoursOnly: true}, 1, 2)
	anthem.Label = "Soldier creatures you control get +1/+2."
	effects.Register(effects.Spec{
		OracleID:     "3fce53fb-6914-4f43-b46b-18ca32ab21b2",
		Name:         "Daru Warchief",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Soldier spells you cast cost {1} less to cast.", effects.YourSpell(), effects.SpellOfCreatureType("Soldier")),
		},
		Static: []game.StaticAbility{anthem},
	})
}
