package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stonybrook Banneret — Creature — Merfolk Wizard {1}{U}, 1/1:
//
//	"Islandwalk
//	 Merfolk spells and Wizard spells you cast cost {1} less to cast."
//
// One reduction (CR 601.2f) over the spells its controller casts that
// are Merfolk or Wizards. Islandwalk is printed data plus the engine's
// landwalk check (game/landwalk.go).
func init() {
	Register(Spec{
		OracleID:        "0d6202c7-b19c-429a-85c4-332024381dd7",
		Name:            "Stonybrook Banneret",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Merfolk spells and Wizard spells you cast cost {1} less to cast.",
				YourSpell(), cmdr2B05SpellOfAnyType("Merfolk", "Wizard")),
		},
	})
}
