package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starnheim Aspirant — Creature — Human Cleric {2}{W}, 2/2:
//
//	"Angel spells you cast cost {2} less to cast."
//
// Generic mana only, applied after any cost increase (CR 601.2f).
func init() {
	Register(Spec{
		OracleID:     "e7b92789-47a9-430a-ad57-be7b6c3d63cb",
		Name:         "Starnheim Aspirant",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Angel spells you cast cost {2} less to cast.",
				YourSpell(), SpellOfCreatureType("Angel")),
		},
	})
}
