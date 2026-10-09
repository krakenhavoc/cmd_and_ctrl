package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geist of Saint Thalia — Legendary Creature — Spirit Cleric {1}{U}, 1/2:
//
//	"Flying
//	 Noncreature spells you cast cost {1} less to cast."
//
// Goblin Electromancer's shape with NoncreatureSpell in place of
// InstantOrSorcerySpell. The reduction spends generic mana only, so it
// never eats a coloured pip.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ef32a4a9-14e2-4738-b4c2-53ce5e1d2a53",
		Name:            "Geist of Saint Thalia",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Noncreature spells you cast cost {1} less to cast.",
				YourSpell(), NoncreatureSpell()),
		},
	})
}
