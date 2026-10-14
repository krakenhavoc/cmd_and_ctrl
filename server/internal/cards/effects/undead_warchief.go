package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undead Warchief — Creature — Zombie {2}{B}{B}, 1/1:
//
//	"Zombie spells you cast cost {1} less to cast.
//	 Zombie creatures you control get +2/+1."
//
// The Zombie Goblin Warchief. The reduction is a cost modifier over
// "Zombie spells you cast" (SpellOfCreatureType, which a changeling
// and a kindred Zombie spell both satisfy), and it spends generic mana
// only. The anthem is not "other": the Warchief is a Zombie creature
// you control, so it pumps itself to a 3/2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6af56bf-bd78-4801-8f6e-033cdd68de3d",
		Name:         "Undead Warchief",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Zombie spells you cast cost {1} less to cast.",
				YourSpell(), SpellOfCreatureType("Zombie")),
		},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Zombie"}, YoursOnly: true}, 2, 1),
		},
	})
}
