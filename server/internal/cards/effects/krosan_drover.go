package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krosan Drover — Creature — Elf {3}{G}, 2/2 (Scourge):
//
//	"Creature spells you cast with mana value 6 or greater cost {2} less
//	 to cast."
//
// One cost modifier (CR 601.2f) over your creature spells whose mana value
// on the stack is 6 or more ({X} counts as the value chosen, CR 202.3e). It
// spends against generic mana only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd8ae400-4f40-42aa-a905-2787de443471",
		Name:         "Krosan Drover",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Creature spells you cast with mana value 6 or greater cost {2} less to cast.",
				YourSpell(), CreatureSpell(), SpellManaValueAtLeast(6)),
		},
	})
}
