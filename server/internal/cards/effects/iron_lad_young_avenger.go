package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Lad, Young Avenger — Legendary Artifact Creature — Human Hero
// {2}{U/R}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Noncreature spells you cast cost {1} less to cast."
//
// The discount is Lyse Hext's CostsLess(1, YourSpell, NoncreatureSpell).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fedcc722-ac5f-464a-bac9-c5f1d4de28d0",
		Name:            "Iron Lad, Young Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Noncreature spells you cast cost {1} less to cast.",
				YourSpell(), NoncreatureSpell()),
		},
	})
}
