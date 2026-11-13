package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	anthem := TribalAnthem(TribeFilter{Tribes: []string{"Soldier"}, YoursOnly: true}, 1, 2)
	anthem.Label = "Soldier creatures you control get +1/+2."
	Register(Spec{
		OracleID:     "3fce53fb-6914-4f43-b46b-18ca32ab21b2",
		Name:         "Daru Warchief",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Soldier spells you cast cost {1} less to cast.", YourSpell(), SpellOfCreatureType("Soldier")),
		},
		Static: []game.StaticAbility{anthem},
	})
}
